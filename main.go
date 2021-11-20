package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/fullstorydev/grpcui/standalone"
	"github.com/sirupsen/logrus"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	userspb "github.com/johanbrandhorst/grpc-postgres/proto"
	"github.com/johanbrandhorst/grpc-postgres/users"
)

const defaultPort = "8080"

func main() {
	log := logrus.New()
	log.Formatter = &logrus.JSONFormatter{}

	pgURL := os.Getenv("POSTGRES_URL")
	if pgURL == "" {
		log.Fatal("POSTGRES_URL must be set")
	}
	parsedURL, err := url.Parse(pgURL)
	if err != nil {
		log.WithError(err).Fatal("Failed to parse POSTGRES_URL")
	}

	port := defaultPort
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.WithError(err).Fatal("Failed to create listener")
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

	var dir userspb.UserServiceServer
	dir, err = users.NewDirectory(log, parsedURL)
	if err != nil {
		log.WithError(err).Fatal("Failed to create user directory")
	}
	userspb.RegisterUserServiceServer(s, dir)

	// Serve gRPC Server
	go func() {
		log.Info("Serving gRPC on ", grpcL.Addr().String())
		sErr := s.Serve(grpcL)
		if sErr != nil {
			log.WithError(err).Fatal("Failed to serve gRPC")
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sAddr := fmt.Sprintf("dns:///localhost:%s", port)
	cc, err := grpc.DialContext(
		ctx,
		sAddr,
		grpc.WithBlock(),
		grpc.WithInsecure(),
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
	log.Info("Serving Web UI on http://localhost:", port)
	err = httpS.Serve(httpL)
	if err != http.ErrServerClosed {
		log.WithError(err).Fatal("Failed to serve Web UI")
	}
}
