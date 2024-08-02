package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/fullstorydev/grpcui/standalone"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	userspb "github.com/johanbrandhorst/grpc-postgres/proto"
	"github.com/johanbrandhorst/grpc-postgres/users"
)

const defaultPort = "8080"

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	pgURL := os.Getenv("POSTGRES_URL")
	if pgURL == "" {
		log.Error("POSTGRES_URL must be set")
		return
	}
	parsedURL, err := url.Parse(pgURL)
	if err != nil {
		log.Error("Failed to parse POSTGRES_URL", "error", err)
		return
	}

	port := defaultPort
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Error("Failed to create listener", "error", err)
		return
	}

	mux := cmux.New(lis)
	grpcL := mux.MatchWithWriters(cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))
	httpL := mux.Match(cmux.Any())

	go func() {
		sErr := mux.Serve()
		if sErr != nil {
			log.Error("Failed to serve cmux", "error", sErr)
			return
		}
	}()

	s := grpc.NewServer()
	reflection.Register(s)

	var dir userspb.UserServiceServer
	dir, err = users.NewDirectory(log, parsedURL)
	if err != nil {
		log.Error("Failed to create user directory", "error", err)
		return
	}
	userspb.RegisterUserServiceServer(s, dir)

	// Serve gRPC Server
	go func() {
		log.Info("Serving gRPC on " + grpcL.Addr().String())
		sErr := s.Serve(grpcL)
		if sErr != nil {
			log.Error("Failed to serve gRPC", "error", sErr)
			return
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sAddr := fmt.Sprintf("dns:///0.0.0.0:%s", port)
	cc, err := grpc.NewClient(sAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("Failed to dial local server", "error", err)
		return
	}
	defer cc.Close()

	handler, err := standalone.HandlerViaReflection(ctx, cc, sAddr)
	if err != nil {
		log.Error("Failed to create grpc UI handler", "error", err)
		return
	}

	httpS := &http.Server{
		Handler: handler,
	}

	// Serve HTTP Server
	log.Info("Serving Web UI on http://0.0.0.0:" + port)
	err = httpS.Serve(httpL)
	if err != http.ErrServerClosed {
		log.Error("failed to serve Web UI", "error", err)
		return
	}
}
