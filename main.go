package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/fullstorydev/grpcui/standalone"
	"github.com/sirupsen/logrus"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"

	userspb "github.com/johanbrandhorst/grpc-postgres/proto"
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
	port = flag.Int("port", 10000, "The server port")
	cert = flag.String("cert", "./insecure/cert.pem", "The path to the server certificate file in PEM format")
	key  = flag.String("key", "./insecure/key.pem", "The path to the server private key in PEM format")
	u    pgURL
)

func main() {
	flag.Var(&u, "postgres-url", "URL formatted address of the postgres server to connect to")
	flag.Parse()

	log := logrus.New()
	log.Formatter = &logrus.TextFormatter{
		TimestampFormat: time.StampMilli,
		FullTimestamp:   true,
	}

	if u.String() == "" {
		log.Fatal("Flag postgres-url is required")
	}

	tlsCert, err := tls.LoadX509KeyPair(*cert, *key)
	if err != nil {
		log.WithError(err).Fatal("Failed to parse certificate and key")
	}
	tlsCert.Leaf, _ = x509.ParseCertificate(tlsCert.Certificate[0]) // Can't fail if LoadX509KeyPair succeeded
	tc := &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
	}
	lis, err := tls.Listen("tcp", fmt.Sprintf(":%d", *port), tc)
	if err != nil {
		log.WithError(err).Fatal("Failed to listen")
	}

	mux := cmux.New(lis)
	grpcL := mux.MatchWithWriters(cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))
	httpL := mux.Match(cmux.Any())

	go func() {
		sErr := mux.Serve()
		if sErr != nil {
			log.WithError(err).Fatal("Failed to serve cmux")
		}
	}()

	s := grpc.NewServer()
	reflection.Register(s)

	var dir userspb.UnstableUserServiceService
	dir, err = users.NewDirectory(log, (*url.URL)(&u))
	if err != nil {
		log.WithError(err).Fatal("Failed to create user directory")
	}
	userspb.RegisterUserServiceService(s, userspb.NewUserServiceService(dir))

	// Serve gRPC Server
	go func() {
		log.Info("Serving gRPC on ", grpcL.Addr().String())
		sErr := s.Serve(grpcL)
		if sErr != nil {
			log.WithError(err).Fatal("Failed to serve gRPC")
		}
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

	handler, err := standalone.HandlerViaReflection(ctx, cc, sAddr)
	if err != nil {
		log.WithError(err).Fatal("Failed to create grpc UI handler")
	}

	httpS := &http.Server{
		Handler: handler,
	}

	// Serve HTTP Server
	log.Info("Serving Web UI on https://localhost:", *port)
	err = httpS.Serve(httpL)
	if err != http.ErrServerClosed {
		log.WithError(err).Fatal("Failed to serve Web UI")
	}
}
