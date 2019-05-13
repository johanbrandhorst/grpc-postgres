package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/fullstorydev/grpcui/standalone"
	"github.com/fullstorydev/grpcurl"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/grpcreflect"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	reflectpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"

	pbUsers "github.com/johanbrandhorst/grpc-postgres/proto"
	"github.com/johanbrandhorst/grpc-postgres/users"
)

type pgURL url.URL

func (p *pgURL) Set(in string) error {
	u, err := url.Parse(in)
	if err != nil {
		return err
	}

	switch u.Scheme {
	case "psql", "postgresql":
	default:
		return errors.New("unexpected scheme in URL")
	}

	*p = pgURL(*u)
	return nil
}

func (p pgURL) String() string {
	return (*url.URL)(&p).String()
}

var (
	port     = flag.Int("port", 10000, "The gRPC server port")
	httpPort = flag.Int("http_port", 11000, "The HTTP UI server port")
	cert     = flag.String("cert", "./insecure/cert.pem", "The path to the server certificate file in PEM format")
	key      = flag.String("key", "./insecure/key.pem", "The path to the server private key in PEM format")
	u        pgURL
)

func main() {
	flag.Var(&u, "postgres-url", "URL formatted address of the postgres server to connect to")
	flag.Parse()

	log := logrus.New()
	log.Formatter = &logrus.TextFormatter{
		TimestampFormat: time.RFC3339,
		FullTimestamp:   true,
	}

	if u.String() == "" {
		log.Fatal("Flag postgres-url is required")
	}

	tlsCert, err := tls.LoadX509KeyPair(*cert, *key)
	if err != nil {
		log.WithError(err).Fatal("Failed to parse certificate and key")
	}
	tlsCert.Leaf, _ = x509.ParseCertificate(tlsCert.Certificate[0])

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.WithError(err).Fatal("Failed to listen")
	}
	s := grpc.NewServer(
		grpc.Creds(credentials.NewServerTLSFromCert(&tlsCert)),
	)
	reflection.Register(s)

	dir, err := users.NewDirectory(log, (*url.URL)(&u))
	if err != nil {
		log.WithError(err).Fatal("Failed to create user directory")
	}
	pbUsers.RegisterUserServiceServer(s, dir)

	// Serve gRPC Server
	go func() {
		log.Info("Serving gRPC on ", lis.Addr().String())
		log.Fatal(s.Serve(lis))
	}()

	cp := x509.NewCertPool()
	cp.AddCert(tlsCert.Leaf)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sAddr := fmt.Sprintf("dns:///localhost:%d", *port)
	cc, err := grpc.DialContext(
		ctx,
		sAddr,
		grpc.WithBlock(),
		grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(cp, "")),
	)
	if err != nil {
		log.WithError(err).Fatal("Failed to dial local server")
	}
	defer cc.Close()

	m, f, err := getMethodsAndFiles(ctx, cc)
	if err != nil {
		log.WithError(err).Fatal("Failed to get server properties")
	}

	httpS := &http.Server{
		Addr:    fmt.Sprintf(":%d", *httpPort),
		Handler: standalone.Handler(cc, sAddr, m, f),
	}

	// Serve HTTP Server
	log.Info("Serving HTTP UI on http://localhost", httpS.Addr)
	log.Fatal(httpS.ListenAndServe())
}

func getMethodsAndFiles(ctx context.Context, cc *grpc.ClientConn) ([]*desc.MethodDescriptor, []*desc.FileDescriptor, error) {
	c := grpcreflect.NewClient(ctx, reflectpb.NewServerReflectionClient(cc))
	source := grpcurl.DescriptorSourceFromServer(ctx, c)

	allServices, err := source.ListServices()
	if err != nil {
		return nil, nil, err
	}

	var m []*desc.MethodDescriptor
	for _, svc := range allServices {
		d, err := source.FindSymbol(svc)
		if err != nil {
			return nil, nil, err
		}
		sd, ok := d.(*desc.ServiceDescriptor)
		if !ok {
			return nil, nil, fmt.Errorf("%s should be a service descriptor but instead is a %T", d.GetFullyQualifiedName(), d)
		}
		if sd.GetFullyQualifiedName() == "grpc.reflection.v1alpha.ServerReflection" {
			continue // skip reflection service
		}
		for _, md := range sd.GetMethods() {
			m = append(m, md)
		}
	}

	f, err := grpcurl.GetAllFiles(source)
	if err != nil {
		return nil, nil, err
	}

	// Release resources
	c.Reset()

	return m, f, nil
}
