package users_test

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/url"
	"runtime"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/ptypes"
	"github.com/google/go-cmp/cmp"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"

	pbUsers "github.com/johanbrandhorst/grpc-postgres/proto"
	"github.com/johanbrandhorst/grpc-postgres/users"
)

func startDatabase(tb testing.TB) *url.URL {
	tb.Helper()

	pgURL := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword("myuser", "mypass"),
		Path:   "mydatabase",
	}
	q := pgURL.Query()
	q.Add("sslmode", "disable")
	pgURL.RawQuery = q.Encode()

	pool, err := dockertest.NewPool("")
	if err != nil {
		tb.Fatalf("Could not connect to docker: %v", err)
	}

	pw, _ := pgURL.User.Password()
	env := []string{
		"POSTGRES_USER=" + pgURL.User.Username(),
		"POSTGRES_PASSWORD=" + pw,
		"POSTGRES_DB=" + pgURL.Path,
	}

	resource, err := pool.Run("postgres", "13-alpine", env)
	if err != nil {
		tb.Fatalf("Could not start postgres container: %v", err)
	}
	tb.Cleanup(func() {
		err = pool.Purge(resource)
		if err != nil {
			tb.Fatalf("Could not purge container: %v", err)
		}
	})

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
		tb.Fatalf("Could not connect to postgres container log output: %v", err)
	}

	tb.Cleanup(func() {
		err = logWaiter.Close()
		if err != nil {
			tb.Fatalf("Could not close container log: %v", err)
		}
		err = logWaiter.Wait()
		if err != nil {
			tb.Fatalf("Could not wait for container log to close: %v", err)
		}
	})

	pool.MaxWait = 10 * time.Second
	err = pool.Retry(func() (err error) {
		db, err := sql.Open("pgx", pgURL.String())
		if err != nil {
			return err
		}
		defer func() {
			cerr := db.Close()
			if err == nil {
				err = cerr
			}
		}()

		return db.Ping()
	})
	if err != nil {
		tb.Fatalf("Could not connect to postgres container: %v", err)
	}

	return pgURL
}

func TestAddDeleteUser(t *testing.T) {
	t.Parallel()

	d, err := users.NewDirectory(logrus.New(), startDatabase(t))
	if err != nil {
		t.Fatalf("Failed to create a new directory: %s", err)
	}
	t.Cleanup(func() {
		err = d.Close()
		if err != nil {
			t.Errorf("Failed to close directory: %s", err)
		}
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	t.Run("When deleting an added user", func(t *testing.T) {
		t.Parallel()

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

		if diff := cmp.Diff(user1, user2, protocmp.Transform()); diff != "" {
			t.Fatalf("Deleted user differed from created user:\n%s", diff)
		}
	})

	t.Run("When using a non-uuid in DeleteUser", func(t *testing.T) {
		t.Parallel()

		_, err = d.DeleteUser(ctx, &pbUsers.DeleteUserRequest{
			Id: "not_a_UUID",
		})
		if status.Code(err) != codes.InvalidArgument {
			t.Fatalf("Did not get correct error when using non-UUID ID in DeleteUser")
		}
	})
}

func TestListUsers(t *testing.T) {
	t.Parallel()

	d, err := users.NewDirectory(logrus.New(), startDatabase(t))
	if err != nil {
		t.Fatalf("Failed to create a new directory: %s", err)
	}
	t.Cleanup(func() {
		err = d.Close()
		if err != nil {
			t.Errorf("Failed to close directory: %s", err)
		}
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

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
		t.Parallel()

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
		t.Parallel()

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
		t.Parallel()

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

	t.Run("Filtering by age and create time", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		srv := NewMockUserService_ListUsersServer(ctrl)
		srv.EXPECT().Context().Return(ctx)
		srv.EXPECT().Send(user2).Return(nil)

		tm, err := ptypes.Timestamp(user2.GetCreateTime())
		if err != nil {
			t.Fatalf("Failed to parse timestamp: %s", err)
		}
		olderThan := time.Since(tm)

		err = d.ListUsers(&pbUsers.ListUsersRequest{
			CreatedSince: user1.GetCreateTime(),
			OlderThan:    ptypes.DurationProto(olderThan),
		}, srv)
		if err != nil {
			t.Fatalf("Failed to list users: %s", err)
		}
		ctrl.Finish()
	})
}
