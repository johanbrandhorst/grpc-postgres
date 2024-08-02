package users_test

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/url"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"

	userspb "github.com/johanbrandhorst/grpc-postgres/proto"
	"github.com/johanbrandhorst/grpc-postgres/users"
)

func startDatabase(tb testing.TB, log *slog.Logger) *url.URL {
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
		OutputStream: os.Stdout,
		ErrorStream:  os.Stdout,
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

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	directory, err := users.NewDirectory(log, startDatabase(t, log))
	if err != nil {
		t.Fatalf("Failed to create a new directory: %s", err)
	}
	t.Cleanup(func() {
		err = directory.Close()
		if err != nil {
			t.Errorf("Failed to close directory: %s", err)
		}
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	t.Run("When deleting an added user", func(t *testing.T) {
		t.Parallel()

		role := userspb.Role_ADMIN
		user1, err := directory.AddUser(ctx, &userspb.AddUserRequest{
			Role: role,
			Name: "Foo",
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

		s := time.Since(user1.CreateTime.AsTime())
		if s > time.Second {
			t.Errorf("CreateTime was longer than 1 second ago: %s", s)
		}

		if user1.GetId() == "" {
			t.Error("Id was not set")
		}

		user2, err := directory.DeleteUser(ctx, &userspb.DeleteUserRequest{
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

		_, err = directory.DeleteUser(ctx, &userspb.DeleteUserRequest{
			Id: "not_a_UUID",
		})
		if status.Code(err) != codes.InvalidArgument {
			t.Fatalf("Did not get correct error when using non-UUID ID in DeleteUser")
		}
	})
}

func TestListUsers(t *testing.T) {
	t.Parallel()

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	directory, err := users.NewDirectory(log, startDatabase(t, log))
	if err != nil {
		t.Fatalf("Failed to create a new directory: %s", err)
	}
	t.Cleanup(func() {
		err = directory.Close()
		if err != nil {
			t.Errorf("Failed to close directory: %s", err)
		}
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	user1, err := directory.AddUser(ctx, &userspb.AddUserRequest{
		Role: userspb.Role_GUEST,
		Name: "Foo",
	})
	if err != nil {
		t.Fatalf("Failed to add a user: %s", err)
	}

	// Sleep so we have slightly different create times
	time.Sleep(500 * time.Millisecond)

	user2, err := directory.AddUser(ctx, &userspb.AddUserRequest{
		Role: userspb.Role_MEMBER,
		Name: "Bar",
	})
	if err != nil {
		t.Fatalf("Failed to add a user: %s", err)
	}

	// Sleep so we have slightly different create times
	time.Sleep(500 * time.Millisecond)

	user3, err := directory.AddUser(ctx, &userspb.AddUserRequest{
		Role: userspb.Role_ADMIN,
		Name: "Baz",
	})
	if err != nil {
		t.Fatalf("Failed to add a user: %s", err)
	}

	t.Run("Returning all users", func(t *testing.T) {
		t.Parallel()

		srv := &listUsersSrvFake{
			ctx: ctx,
		}

		err := directory.ListUsers(new(userspb.ListUsersRequest), srv)
		if err != nil {
			t.Fatalf("Failed to list users: %s", err)
		}

		if len(srv.users) != 3 {
			t.Fatal("Did not receive 3 users as expected")
		}
		if diff := cmp.Diff(srv.users[0], user1, protocmp.Transform()); diff != "" {
			t.Errorf("First user didn't match user1: %s", diff)
		}
		if diff := cmp.Diff(srv.users[1], user2, protocmp.Transform()); diff != "" {
			t.Errorf("Second user didn't match user2: %s", diff)
		}
		if diff := cmp.Diff(srv.users[2], user3, protocmp.Transform()); diff != "" {
			t.Errorf("Third user didn't match user3: %s", diff)
		}
	})

	t.Run("Filtering by age", func(t *testing.T) {
		t.Parallel()

		srv := &listUsersSrvFake{
			ctx: ctx,
		}

		olderThan := time.Since(user2.GetCreateTime().AsTime())

		err := directory.ListUsers(&userspb.ListUsersRequest{
			OlderThan: durationpb.New(olderThan),
		}, srv)
		if err != nil {
			t.Fatalf("Failed to list users: %s", err)
		}

		if len(srv.users) != 2 {
			t.Fatal("Did not receive 2 users as expected")
		}
		if diff := cmp.Diff(srv.users[0], user1, protocmp.Transform()); diff != "" {
			t.Errorf("First user didn't match user1: %s", diff)
		}
		if diff := cmp.Diff(srv.users[1], user2, protocmp.Transform()); diff != "" {
			t.Errorf("Second user didn't match user2: %s", diff)
		}
	})

	t.Run("Filtering by create time", func(t *testing.T) {
		t.Parallel()

		srv := &listUsersSrvFake{
			ctx: ctx,
		}

		err := directory.ListUsers(&userspb.ListUsersRequest{
			CreatedSince: user1.GetCreateTime(),
		}, srv)
		if err != nil {
			t.Fatalf("Failed to list users: %s", err)
		}

		if len(srv.users) != 2 {
			t.Fatal("Did not receive 2 users as expected")
		}
		if diff := cmp.Diff(srv.users[0], user2, protocmp.Transform()); diff != "" {
			t.Errorf("First user didn't match user2: %s", diff)
		}
		if diff := cmp.Diff(srv.users[1], user3, protocmp.Transform()); diff != "" {
			t.Errorf("Second user didn't match user3: %s", diff)
		}
	})

	t.Run("Filtering by age and create time", func(t *testing.T) {
		t.Parallel()

		srv := &listUsersSrvFake{
			ctx: ctx,
		}

		olderThan := time.Since(user2.GetCreateTime().AsTime())

		err := directory.ListUsers(&userspb.ListUsersRequest{
			CreatedSince: user1.GetCreateTime(),
			OlderThan:    durationpb.New(olderThan),
		}, srv)
		if err != nil {
			t.Fatalf("Failed to list users: %s", err)
		}
		if len(srv.users) != 1 {
			t.Fatal("Did not receive 2 users as expected")
		}
		if diff := cmp.Diff(srv.users[0], user2, protocmp.Transform()); diff != "" {
			t.Errorf("First user didn't match user2: %s", diff)
		}
	})
}

func TestAddUsers(t *testing.T) {
	t.Parallel()

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	directory, err := users.NewDirectory(log, startDatabase(t, log))
	if err != nil {
		t.Fatalf("Failed to create a new directory: %s", err)
	}
	t.Cleanup(func() {
		err = directory.Close()
		if err != nil {
			t.Errorf("Failed to close directory: %s", err)
		}
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	numUsers := 10
	t.Run(fmt.Sprintf("Add %d users", numUsers), func(t *testing.T) {
		t.Parallel()

		addSrv := &addUsersSrvFake{
			ctx: ctx,
		}
		for i := 0; i < numUsers; i++ {
			addSrv.reqs = append(addSrv.reqs, &userspb.AddUserRequest{
				Role: userspb.Role_MEMBER,
				Name: "Foo",
			})
		}

		err = directory.AddUsers(addSrv)
		if err != nil {
			t.Fatalf("Failed to add users: %s", err)
		}

		listSrv := &listUsersSrvFake{
			ctx: ctx,
		}
		err = directory.ListUsers(new(userspb.ListUsersRequest), listSrv)
		if err != nil {
			t.Fatalf("Failed to list users: %s", err)
		}
		if len(listSrv.users) != numUsers {
			t.Fatalf("Expected %d users, got %d", numUsers, len(listSrv.users))
		}
	})
}

func BenchmarkAddUsers(b *testing.B) {
	b.Skip("Benchmarks take a while to run")
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	directory, err := users.NewDirectory(log, startDatabase(b, log))
	if err != nil {
		b.Fatalf("Failed to create a new directory: %s", err)
	}
	b.Cleanup(func() {
		err = directory.Close()
		if err != nil {
			b.Errorf("Failed to close directory: %s", err)
		}
	})

	ctx, cancel := context.WithCancel(context.Background())
	b.Cleanup(cancel)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		addSrv := &addUsersSrvFake{
			ctx: ctx,
		}
		for pb.Next() {
			addSrv.reqs = append(addSrv.reqs, &userspb.AddUserRequest{
				Role: userspb.Role_MEMBER,
				Name: "Foo",
			})
		}

		err = directory.AddUsers(addSrv)
		if err != nil {
			b.Fatalf("Failed to add users: %s", err)
		}
	})
}

var benchmarkUser *userspb.User

func BenchmarkAddUser(b *testing.B) {
	b.Skip("Benchmarks take a while to run")
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	directory, err := users.NewDirectory(log, startDatabase(b, log))
	if err != nil {
		b.Fatalf("Failed to create a new directory: %s", err)
	}
	b.Cleanup(func() {
		err = directory.Close()
		if err != nil {
			b.Errorf("Failed to close directory: %s", err)
		}
	})

	ctx, cancel := context.WithCancel(context.Background())
	b.Cleanup(cancel)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var user *userspb.User
		for pb.Next() {
			user, err = directory.AddUser(ctx, &userspb.AddUserRequest{
				Role: userspb.Role_MEMBER,
				Name: "Foo",
			})
			if err != nil {
				b.Fatalf("Failed to add user: %s", err)
			}
		}
		// Avoid compiler optimizing out the loop
		benchmarkUser = user
	})
}

type addUsersSrvFake struct {
	grpc.ServerStream
	reqs []*userspb.AddUserRequest
	ctx  context.Context
}

func (a *addUsersSrvFake) SendAndClose(_ *emptypb.Empty) error {
	return nil
}

func (a *addUsersSrvFake) Recv() (*userspb.AddUserRequest, error) {
	if len(a.reqs) == 0 {
		return nil, io.EOF
	}
	// Pop a request off the top
	req := a.reqs[0]
	a.reqs = a.reqs[1:]
	return req, nil
}

// Context returns the context for this stream.
func (a *addUsersSrvFake) Context() context.Context {
	return a.ctx
}

type listUsersSrvFake struct {
	grpc.ServerStream
	ctx   context.Context
	users []*userspb.User
}

func (l *listUsersSrvFake) Send(user *userspb.User) error {
	l.users = append(l.users, user)
	return nil
}

// Context returns the context for this stream.
func (l *listUsersSrvFake) Context() context.Context {
	return l.ctx
}
