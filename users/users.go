package users

import (
	"context"
	"database/sql"
	"net/url"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/log/logrusadapter"
	"github.com/jackc/pgx/stdlib"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pbUsers "github.com/johanbrandhorst/grpc-postgres/proto"
)

// Directory stores a directory of users.
type Directory struct {
	logger *logrus.Logger
	db     *sql.DB
	sb     squirrel.StatementBuilderType
}

// NewDirectory creates a new Directory, connecting it to the postgres server on
// the URL provided.
func NewDirectory(logger *logrus.Logger, pgURL *url.URL) (*Directory, error) {
	c, err := pgx.ParseURI((pgURL.String()))
	if err != nil {
		return nil, err
	}

	c.Logger = logrusadapter.NewLogger(logger)
	db := stdlib.OpenDB(c)

	err = validateSchema(db)
	if err != nil {
		return nil, err
	}

	return &Directory{
		logger: logger,
		db:     db,
		sb:     squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).RunWith(db),
	}, nil
}

// AddUser adds a user to the directory
func (d Directory) AddUser(ctx context.Context, req *pbUsers.AddUserRequest) (*pbUsers.User, error) {
	q := d.sb.Insert(
		"users",
	).SetMap(map[string]interface{}{
		"role": (roleWrapper)(req.GetRole()),
	}).Suffix(
		"RETURNING id, role, create_time",
	)

	return scanUser(q.QueryRowContext(ctx))
}

// ListUsers lists users in the directory, subject to the request filters.
func (d Directory) ListUsers(req *pbUsers.ListUsersRequest, srv pbUsers.UserService_ListUsersServer) (err error) {
	q := d.sb.Select(
		"id",
		"role",
		"create_time",
	).From(
		"users",
	)

	if req.GetCreatedSince() != nil {
		q = q.Where(squirrel.Gt{
			"create_time": (*timeWrapper)(req.GetCreatedSince()),
		})
	}

	if req.GetOlderThan() != nil {
		q = q.Where(
			squirrel.Expr(
				"CURRENT_TIMESTAMP - create_time > $1", (*durationWrapper)(req.GetOlderThan()),
			),
		)
	}

	defer func() {
		d.logger.WithError(err).Debug("stuff")
	}()

	rows, err := q.QueryContext(srv.Context())
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	defer func() {
		cerr := rows.Close()
		if err == nil && cerr != nil {
			err = status.Error(codes.Internal, cerr.Error())
		}
	}()

	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		err = srv.Send(user)
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}
	}

	err = rows.Err()
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	return nil
}
