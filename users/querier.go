// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.18.0

package users

import (
	"context"

	"github.com/jackc/pgtype"
)

type Querier interface {
	AddUser(ctx context.Context, arg AddUserParams) (User, error)
	DeleteUser(ctx context.Context, id pgtype.UUID) (User, error)
}

var _ Querier = (*Queries)(nil)
