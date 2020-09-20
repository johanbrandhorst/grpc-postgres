package users

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/log/logrusadapter"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	userspb "github.com/johanbrandhorst/grpc-postgres/proto"
)

// Directory stores a directory of users.
type Directory struct {
	logger  *logrus.Logger
	db      *sql.DB
	sb      squirrel.StatementBuilderType
	querier Querier
}

// NewDirectory creates a new Directory, connecting it to the postgres server on
// the URL provided.
func NewDirectory(logger *logrus.Logger, pgURL *url.URL) (*Directory, error) {
	c, err := pgx.ParseConfig(pgURL.String())
	if err != nil {
		return nil, fmt.Errorf("parsing postgres URI: %w", err)
	}

	c.Logger = logrusadapter.NewLogger(logger)
	db := stdlib.OpenDB(*c)

	err = validateSchema(db)
	if err != nil {
		return nil, fmt.Errorf("validating schema: %w", err)
	}

	return &Directory{
		logger:  logger,
		db:      db,
		sb:      squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).RunWith(db),
		querier: New(db),
	}, nil
}

// Close releases any resources.
func (d Directory) Close() error {
	return d.db.Close()
}

// AddUser adds a user to the directory.
func (d Directory) AddUser(ctx context.Context, req *userspb.AddUserRequest) (*userspb.User, error) {
	pgRole, err := roleProtoToPostgres(req.Role)
	if err != nil {
		return nil, err
	}
	pgUser, err := d.querier.AddUser(ctx, AddUserParams{
		Role: pgRole,
		Name: req.Name,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unexpected error adding user: %s", err.Error())
	}
	return userPostgresToProto(pgUser)
}

// AddUsers adds a large amount of users efficiently.
func (d Directory) AddUsers(srv userspb.UserService_AddUsersServer) (retErr error) {
	conn, err := d.db.Conn(srv.Context())
	if err != nil {
		status.Errorf(codes.Internal, "unexpected error getting connection: %s", err.Error())
	}
	defer func() {
		err := conn.Close()
		if retErr == nil {
			retErr = err
		}
	}()
	err = conn.Raw(func(driverConn interface{}) error {
		conn := driverConn.(*stdlib.Conn).Conn()
		// CopyFrom uses the Postgres COPY protocol to perform bulk data insertion.
		// CopyFrom can be faster than an insert with as few as 5 rows.
		_, err = conn.CopyFrom(
			srv.Context(),
			pgx.Identifier{"users"},
			[]string{"role", "name"},
			&usersSource{
				getUser: srv.Recv,
			},
		)
		if err != nil {
			return status.Errorf(codes.Internal, "unexpected error inserting users: %s", err.Error())
		}
		return nil
	})
	if err != nil {
		return err
	}
	return srv.SendAndClose(new(emptypb.Empty))
}

// DeleteUser deletes the user, if found.
func (d Directory) DeleteUser(ctx context.Context, req *userspb.DeleteUserRequest) (*userspb.User, error) {
	var userID pgtype.UUID
	err := userID.Set(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid UUID provided")
	}
	pgUser, err := d.querier.DeleteUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return userPostgresToProto(pgUser)
}

// ListUsers lists users in the directory, subject to the request filters.
func (d Directory) ListUsers(req *userspb.ListUsersRequest, srv userspb.UserService_ListUsersServer) (retErr error) {
	q := d.sb.Select(
		"id",
		"role",
		"create_time",
		"name",
	).From(
		"users",
	).OrderBy(
		"create_time ASC",
	)

	if req.GetCreatedSince() != nil {
		var pgTime pgtype.Timestamptz
		err := pgTime.Set(req.GetCreatedSince().AsTime())
		if err != nil {
			return status.Errorf(codes.InvalidArgument, "invalid timestamp: %s", err.Error())
		}
		q = q.Where(squirrel.Gt{
			"create_time": pgTime,
		})
	}

	if req.GetOlderThan() != nil {
		var pgInterval pgtype.Interval
		err := pgInterval.Set(req.GetOlderThan().AsDuration())
		if err != nil {
			return status.Errorf(codes.InvalidArgument, "invalid duration: %s", err.Error())
		}
		q = q.Where(
			squirrel.Expr(
				"CURRENT_TIMESTAMP - create_time > ?", pgInterval,
			),
		)
	}

	rows, retErr := q.QueryContext(srv.Context())
	if retErr != nil {
		return status.Error(codes.Internal, retErr.Error())
	}
	defer func() {
		cerr := rows.Close()
		if retErr == nil && cerr != nil {
			retErr = status.Error(codes.Internal, cerr.Error())
		}
	}()

	for rows.Next() {
		var pgUser User
		err := rows.Scan(
			&pgUser.ID,
			&pgUser.Role,
			&pgUser.CreateTime,
			&pgUser.Name,
		)
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}
		protoUser, err := userPostgresToProto(pgUser)
		if err != nil {
			return err
		}
		err = srv.Send(protoUser)
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}
	}

	retErr = rows.Err()
	if retErr != nil {
		return status.Error(codes.Internal, retErr.Error())
	}

	return nil
}
