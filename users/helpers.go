package users

import (
	"database/sql"

	migrate "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	bindata "github.com/golang-migrate/migrate/v4/source/go_bindata"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pbUsers "github.com/johanbrandhorst/grpc-postgres/proto"
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

func userPostgresToProto(pgUser User) (*pbUsers.User, error) {
	protoRole, err := rolePostgresToProto(pgUser.Role)
	if err != nil {
		return nil, err
	}
	var userID string
	err = pgUser.ID.AssignTo(&userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to assign UUID to string: %v", err)
	}
	return &pbUsers.User{
		CreateTime: timestamppb.New(pgUser.CreateTime),
		Id:         userID,
		Role:       protoRole,
	}, nil
}

func rolePostgresToProto(pgRole Role) (pbUsers.Role, error) {
	switch pgRole {
	case RoleGuest:
		return pbUsers.Role_GUEST, nil
	case RoleAdmin:
		return pbUsers.Role_ADMIN, nil
	case RoleMember:
		return pbUsers.Role_MEMBER, nil
	default:
		return 0, status.Errorf(codes.Internal, "unknown role type %q", pgRole)
	}
}

func roleProtoToPostgres(pbRole pbUsers.Role) (Role, error) {
	switch pbRole {
	case pbUsers.Role_GUEST:
		return RoleGuest, nil
	case pbUsers.Role_ADMIN:
		return RoleAdmin, nil
	case pbUsers.Role_MEMBER:
		return RoleMember, nil
	default:
		return "", status.Errorf(codes.InvalidArgument, "unknown role type %q", pbRole)
	}
}
