package users

import (
	"database/sql/driver"
	"fmt"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/jackc/pgtype"

	pbUsers "github.com/johanbrandhorst/grpc-postgres/proto"
)

type roleWrapper pbUsers.Role

// Value implements database/sql/driver.Valuer for pbUsers.Role
func (rw roleWrapper) Value() (driver.Value, error) {
	switch pbUsers.Role(rw) {
	case pbUsers.Role_GUEST:
		return "guest", nil
	case pbUsers.Role_MEMBER:
		return "member", nil
	case pbUsers.Role_ADMIN:
		return "admin", nil
	default:
		return nil, fmt.Errorf("invalid Role: %q", rw)
	}
}

// Scan implements database/sql/driver.Scanner for pbUsers.Role
func (rw *roleWrapper) Scan(in interface{}) error {
	switch in.(string) {
	case "guest":
		*rw = roleWrapper(pbUsers.Role_GUEST)
		return nil
	case "member":
		*rw = roleWrapper(pbUsers.Role_MEMBER)
		return nil
	case "admin":
		*rw = roleWrapper(pbUsers.Role_ADMIN)
		return nil
	default:
		return fmt.Errorf("invalid Role: %q", in.(string))
	}
}

type timeWrapper timestamp.Timestamp

// Value implements database/sql/driver.Valuer for timestamp.Timestamp
func (tw *timeWrapper) Value() (driver.Value, error) {
	return ptypes.Timestamp((*timestamp.Timestamp)(tw))
}

// Scan implements database/sql/driver.Scanner for timestamp.Timestamp
func (tw *timeWrapper) Scan(in interface{}) error {
	var t pgtype.Timestamptz
	err := t.Scan(in)
	if err != nil {
		return err
	}

	*tw = timeWrapper(timestamp.Timestamp{
		Seconds: t.Time.Unix(),
		Nanos:   int32(t.Time.Nanosecond()),
	})

	return nil
}

type durationWrapper duration.Duration

// Value implements database/sql/driver.Valuer for duration.Duration
func (dw *durationWrapper) Value() (driver.Value, error) {
	d, err := ptypes.Duration((*duration.Duration)(dw))
	if err != nil {
		return nil, err
	}

	i := pgtype.Interval{
		Microseconds: int64(d) / 1000,
		Status:       pgtype.Present,
	}

	return i.Value()
}
