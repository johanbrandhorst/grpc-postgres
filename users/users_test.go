package users_test

import (
	"context"
	"database/sql"
	"net"
	"net/url"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/ptypes"
	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
	"github.com/sirupsen/logrus"

	pbUsers "github.com/johanbrandhorst/grpc-postgres/proto"
	"github.com/johanbrandhorst/grpc-postgres/users"
)

var (
	log *logrus.Logger

	pgURL *url.URL
)

func TestMain(m *testing.M) {
	code := 0
	defer func() {
		os.Exit(code)
	}()

	log = logrus.New()
	log.Formatter = &logrus.TextFormatter{
		TimestampFormat: time.RFC3339,
		FullTimestamp:   true,
		ForceColors:     true,
	}

	pgURL = &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword("myuser", "mypass"),
		Path:   "mydatabase",
	}
	q := pgURL.Query()
	q.Add("sslmode", "disable")
	pgURL.RawQuery = q.Encode()

	pool, err := dockertest.NewPool("")
	if err != nil {
		log.WithError(err).Fatal("Could not connect to docker")
	}

	pw, _ := pgURL.User.Password()
	runOpts := dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "latest",
		Env: []string{
			"POSTGRES_USER=" + pgURL.User.Username(),
			"POSTGRES_PASSWORD=" + pw,
			"POSTGRES_DB=" + pgURL.Path,
		},
	}

	resource, err := pool.RunWithOptions(&runOpts)
	if err != nil {
		log.WithError(err).Fatal("Could start postgres container")
	}
	defer func() {
		err = pool.Purge(resource)
		if err != nil {
			log.WithError(err).Error("Could not purge resource")
		}
	}()

	pgURL.Host = resource.Container.NetworkSettings.IPAddress

	// Docker layer network is different on Mac
	if runtime.GOOS == "darwin" {
		pgURL.Host = net.JoinHostPort(resource.GetBoundIP("5432/tcp"), resource.GetPort("5432/tcp"))
	}

	logWaiter, err := pool.Client.AttachToContainerNonBlocking(docker.AttachToContainerOptions{
		Container:    resource.Container.ID,
		OutputStream: log.Writer(),
		ErrorStream:  log.Writer(),
		Stderr:       true,
		Stdout:       true,
		Stream:       true,
	})
	if err != nil {
		log.WithError(err).Fatal("Could not connect to postgres container log output")
	}
	defer func() {
		err = logWaiter.Close()
		if err != nil {
			log.WithError(err).Error("Could not close container log")
		}
		err = logWaiter.Wait()
		if err != nil {
			log.WithError(err).Error("Could not wait for container log to close")
		}
	}()

	pool.MaxWait = 10 * time.Second
	err = pool.Retry(func() error {
		db, err := sql.Open("postgres", pgURL.String())
		if err != nil {
			return err
		}
		return db.Ping()
	})
	if err != nil {
		log.WithError(err).Fatal("Could not connect to postgres server")
	}

	code = m.Run()
}

func TestAddDeleteUser(t *testing.T) {
	d, err := users.NewDirectory(log, pgURL)
	if err != nil {
		t.Fatalf("Failed to create a new directory: %s", err)
	}
	defer func() {
		err = d.Close()
		if err != nil {
			t.Errorf("Failed to close directory: %s", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	role := pbUsers.Role_ADMIN
	user1, err := d.AddUser(ctx, &pbUsers.AddUserRequest{
		Role: role,
	})
	if err != nil {
		t.Fatalf("Failed to add a user: %s", err)
	}

	if user1.GetRole() != role {
		t.Errorf("Got role %q, wanted role %q", user1.GetRole(), role)
	}
	if user1.GetCreateTime() == nil {
		t.Fatal("CreateTime was not set")
	}

	tm, err := ptypes.Timestamp(user1.GetCreateTime())
	if err != nil {
		t.Fatalf("CreateTime could not be parsed: %s", err)
	}

	s := time.Since(tm)
	if s > time.Second {
		t.Errorf("CreateTime was longer than 1 second ago: %s", s)
	}

	if user1.GetId() == "" {
		t.Error("Id was not set")
	}

	user2, err := d.DeleteUser(ctx, &pbUsers.DeleteUserRequest{
		Id: user1.GetId(),
	})
	if err != nil {
		t.Fatalf("Failed to delete user: %s", err)
	}

	if user1.GetRole() != user2.GetRole() ||
		user1.GetId() != user2.GetId() ||
		user1.GetCreateTime().GetNanos() != user2.GetCreateTime().GetNanos() ||
		user1.GetCreateTime().GetSeconds() != user2.GetCreateTime().GetSeconds() {
		t.Fatalf("Deleted user differed from created user:\n%#v\n%#v", user1, user2)
	}
}

func TestListUsers(t *testing.T) {
	d, err := users.NewDirectory(log, pgURL)
	if err != nil {
		t.Fatalf("Failed to create a new directory: %s", err)
	}
	defer func() {
		err = d.Close()
		if err != nil {
			t.Errorf("Failed to close directory: %s", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	user1, err := d.AddUser(ctx, &pbUsers.AddUserRequest{
		Role: pbUsers.Role_GUEST,
	})
	if err != nil {
		t.Fatalf("Failed to add a user: %s", err)
	}

	// Sleep so we have slightly different create times
	time.Sleep(500 * time.Millisecond)

	user2, err := d.AddUser(ctx, &pbUsers.AddUserRequest{
		Role: pbUsers.Role_MEMBER,
	})
	if err != nil {
		t.Fatalf("Failed to add a user: %s", err)
	}

	// Sleep so we have slightly different create times
	time.Sleep(500 * time.Millisecond)

	user3, err := d.AddUser(ctx, &pbUsers.AddUserRequest{
		Role: pbUsers.Role_ADMIN,
	})
	if err != nil {
		t.Fatalf("Failed to add a user: %s", err)
	}

	t.Run("Returning all users", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		srv := NewMockUserService_ListUsersServer(ctrl)
		srv.EXPECT().Context().Return(ctx)
		srv.EXPECT().Send(user1).Return(nil)
		srv.EXPECT().Send(user2).Return(nil)
		srv.EXPECT().Send(user3).Return(nil)

		err = d.ListUsers(&pbUsers.ListUsersRequest{}, srv)
		if err != nil {
			t.Fatalf("Failed to list users: %s", err)
		}
		ctrl.Finish()
	})

	t.Run("Filtering by age", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		srv := NewMockUserService_ListUsersServer(ctrl)
		srv.EXPECT().Context().Return(ctx)
		srv.EXPECT().Send(user1).Return(nil)
		srv.EXPECT().Send(user2).Return(nil)

		tm, err := ptypes.Timestamp(user2.GetCreateTime())
		if err != nil {
			t.Fatalf("Failed to parse timestamp: %s", err)
		}
		olderThan := time.Since(tm)

		err = d.ListUsers(&pbUsers.ListUsersRequest{
			OlderThan: ptypes.DurationProto(olderThan),
		}, srv)
		if err != nil {
			t.Fatalf("Failed to list users: %s", err)
		}
		ctrl.Finish()
	})

	t.Run("Filtering by create time", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		srv := NewMockUserService_ListUsersServer(ctrl)
		srv.EXPECT().Context().Return(ctx)
		srv.EXPECT().Send(user2).Return(nil)
		srv.EXPECT().Send(user3).Return(nil)

		err = d.ListUsers(&pbUsers.ListUsersRequest{
			CreatedSince: user1.GetCreateTime(),
		}, srv)
		if err != nil {
			t.Fatalf("Failed to list users: %s", err)
		}
		ctrl.Finish()
	})
}
