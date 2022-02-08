package users

import (
	"embed"
	"io"

	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	userspb "github.com/johanbrandhorst/grpc-postgres/proto"
)

//go:embed migrations/*.sql
var fs embed.FS

// version defines the current migration version. This ensures the app
// is always compatible with the version of the database.
const version = 1


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
		Name:       pgUser.Name,
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
	getUser    func() (*userspb.AddUserRequest, error)
	nextValues []interface{}
	err        error
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
	u.nextValues = []interface{}{pgRole, req.Name}
	return true
}

func (u *usersSource) Values() ([]interface{}, error) {
	return u.nextValues, nil
}

func (u *usersSource) Err() error {
	if u.err == io.EOF {
		// This is actually success, so we don't want to return an error
		return nil
	}
	return u.err
}
