package main

import (
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"

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
		TimestampFormat: time.RFC3339,
		FullTimestamp:   true,
	}

	if u.String() == "" {
		log.Fatal("Flag postgres-url is required")
	}

	cert, err := tls.LoadX509KeyPair(*cert, *key)
	if err != nil {
		log.WithError(err).Fatal("Failed to parse certificate and key")
	}

	addr := fmt.Sprintf(":%d", *port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.WithError(err).Fatal("Failed to listen")
	}
	s := grpc.NewServer(
		grpc.Creds(credentials.NewServerTLSFromCert(&cert)),
	)
	reflection.Register(s)

	dir, err := users.NewDirectory(log, (*url.URL)(&u))
	if err != nil {
		log.WithError(err).Fatal("Failed to create user directory")
	}
	pbUsers.RegisterUserServiceServer(s, dir)

	// Serve gRPC Server
	log.Info("Serving gRPC on https://", addr)
	log.Fatal(s.Serve(lis))
}
