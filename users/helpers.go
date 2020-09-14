package users

import (
	"database/sql"
	"io"

	migrate "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	bindata "github.com/golang-migrate/migrate/v4/source/go_bindata"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	userspb "github.com/johanbrandhorst/grpc-postgres/proto"
	"github.com/johanbrandhorst/grpc-postgres/users/migrations"
)

// version defines the current migration version. This ensures the app
// is always compatible with the version of the database.
const version = 1

// Migrate migrates the Postgres schema to the current version.
func validateSchema(db *sql.DB) error {
	sourceInstance, err := bindata.WithInstance(bindata.Resource(migrations.AssetNames(), migrations.Asset))
	if err != nil {
		return err
	}
	targetInstance, err := postgres.WithInstance(db, new(postgres.Config))
	if err != nil {
		return err
	}
	m, err := migrate.NewWithInstance("go-bindata", sourceInstance, "postgres", targetInstance)
	if err != nil {
		return err
	}
	err = m.Migrate(version) // current version
	if err != nil && err != migrate.ErrNoChange {
		return err
	}
	return sourceInstance.Close()
}

func userPostgresToProto(pgUser User) (*userspb.User, error) {
	protoRole, err := rolePostgresToProto(pgUser.Role)
	if err != nil {
		return nil, err
	}
	var userID string
	err = pgUser.ID.AssignTo(&userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to assign UUID to string: %s", err.Error())
	}
	return &userspb.User{
		CreateTime: timestamppb.New(pgUser.CreateTime),
		Id:         userID,
		Role:       protoRole,
	}, nil
}

func userProtoToPostgres(protoUser *userspb.User) (User, error) {
	pgRole, err := roleProtoToPostgres(protoUser.Role)
	if err != nil {
		return User{}, err
	}
	var userID pgtype.UUID
	err = userID.Set(protoUser.Id)
	if err != nil {
		return User{}, status.Errorf(codes.Internal, "failed to parse user ID as UUID: %s", err.Error())
	}
	return User{
		ID:         userID,
		CreateTime: protoUser.CreateTime.AsTime(),
		Role:       pgRole,
	}, nil
}

func rolePostgresToProto(pgRole Role) (userspb.Role, error) {
	switch pgRole {
	case RoleGuest:
		return userspb.Role_GUEST, nil
	case RoleAdmin:
		return userspb.Role_ADMIN, nil
	case RoleMember:
		return userspb.Role_MEMBER, nil
	default:
		return 0, status.Errorf(codes.Internal, "unknown role type %q", pgRole)
	}
}

func roleProtoToPostgres(pbRole userspb.Role) (Role, error) {
	switch pbRole {
	case userspb.Role_GUEST:
		return RoleGuest, nil
	case userspb.Role_ADMIN:
		return RoleAdmin, nil
	case userspb.Role_MEMBER:
		return RoleMember, nil
	default:
		return "", status.Errorf(codes.InvalidArgument, "unknown role type %q", pbRole)
	}
}

var _ pgx.CopyFromSource = (*usersSource)(nil)

type usersSource struct {
	getUser   func() (*userspb.AddUserRequest, error)
	nextValue interface{}
	err       error
}

func (u *usersSource) Next() bool {
	if u.err != nil {
		return false
	}
	var req *userspb.AddUserRequest
	req, u.err = u.getUser()
	if u.err != nil {
		return false
	}
	var pgRole Role
	pgRole, u.err = roleProtoToPostgres(req.Role)
	if u.err != nil {
		return false
	}
	u.nextValue = pgRole
	return true
}

func (u *usersSource) Values() ([]interface{}, error) {
	return []interface{}{u.nextValue}, nil
}

func (u *usersSource) Err() error {
	if u.err == io.EOF {
		// This is actually success, so we don't want to return an error
		return nil
	}
	return u.err
}
