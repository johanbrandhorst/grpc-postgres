package main

import (
	"context"
	"flag"
	"io"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/types/known/durationpb"

	pbUsers "github.com/johanbrandhorst/grpc-postgres/proto"
)

var (
	addr      = flag.String("addr", "dns:///localhost:10000", "The address of the gRPC server")
	cert      = flag.String("cert", "../insecure/cert.pem", "The path of the server certificate")
	olderThan = flag.Duration("older_than", 0, "Filter to use when listing users.")
	add       = flag.Bool("add", false, "Whether to add another user")
)

func main() {
	flag.Parse()

	log := logrus.New()
	log.Formatter = &logrus.TextFormatter{
		TimestampFormat: time.RFC3339,
		FullTimestamp:   true,
	}

	creds, err := credentials.NewClientTLSFromFile(*cert, "")
	if err != nil {
		log.WithError(err).Fatal("Failed to create server credentials")
	}

	conn, err := grpc.Dial(
		*addr,
		grpc.WithTransportCredentials(creds),
	)
	if err != nil {
		log.WithError(err).Fatal("Failed to dial the server")
	}

	c := pbUsers.NewUserServiceClient(conn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if *add {
		user, err := c.AddUser(ctx, &pbUsers.AddUserRequest{
			Role: pbUsers.Role_GUEST,
		})
		if err != nil {
			log.WithError(err).Fatal("Failed to add user")
		}
		log.WithFields(logrus.Fields{
			"id":          user.GetId(),
			"role":        user.GetRole().String(),
			"create_time": user.GetCreateTime().AsTime().Local().Format(time.RFC3339),
		}).Info("Added user")
	}

	lReq := new(pbUsers.ListUsersRequest)

	if *olderThan != 0 {
		lReq.OlderThan = durationpb.New(*olderThan)
	}

	srv, err := c.ListUsers(ctx, lReq)
	if err != nil {
		log.WithError(err).Fatal("Failed to list users")
	}

	for {
		user, err := srv.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.WithError(err).Fatal("Error while receiving users")
		}

		log.WithFields(logrus.Fields{
			"id":          user.GetId(),
			"role":        user.GetRole().String(),
			"create_time": user.GetCreateTime().AsTime().Local().Format(time.RFC3339),
		}).Info("Read user")
	}

	log.Info("Finished")
}
