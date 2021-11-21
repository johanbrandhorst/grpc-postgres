package main

import (
	"context"
	"crypto/tls"
	"flag"
	"io"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/types/known/durationpb"

	userspb "github.com/johanbrandhorst/grpc-postgres/proto"
)

var (
	addr      = flag.String("addr", "dns:///localhost:10000", "The address of the gRPC server")
	olderThan = flag.Duration("older_than", 0, "Filter to use when listing users.")
	add       = flag.Bool("add", false, "Whether to add another user")
	insecure  = flag.Bool("insecure", false, "Whether to use insecure TLS")
)

func main() {
	flag.Parse()

	log := logrus.New()
	log.Formatter = &logrus.TextFormatter{
		TimestampFormat: time.RFC3339,
		FullTimestamp:   true,
	}

	var opts []grpc.DialOption
	if *insecure {
		opts = append(opts, grpc.WithInsecure())
	} else {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	}

	conn, err := grpc.Dial(*addr, opts...)
	if err != nil {
		log.WithError(err).Fatal("Failed to dial the server")
	}

	c := userspb.NewUserServiceClient(conn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if *add {
		user, err := c.AddUser(ctx, &userspb.AddUserRequest{
			Role: userspb.Role_GUEST,
			Name: "Foo",
		})
		if err != nil {
			log.WithError(err).Fatal("Failed to add user")
		}
		log.WithFields(logrus.Fields{
			"id":          user.GetId(),
			"role":        user.GetRole().String(),
			"create_time": user.GetCreateTime().AsTime().Local().Format(time.RFC3339),
			"name":        user.GetName(),
		}).Info("Added user")
	}

	lReq := new(userspb.ListUsersRequest)

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
			"name":        user.GetName(),
		}).Info("Read user")
	}

	log.Info("Finished")
}
