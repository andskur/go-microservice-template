package models

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/gofrs/uuid"
)

// User represents a user in the system.
//
//nolint:govet // field alignment kept for readability and conventional ordering
type User struct {
	tableName struct{} `pg:"users,discard_unknown_columns"` //nolint:unused // go-pg table marker

	UUID      uuid.UUID  `pg:"uuid,pk,type:uuid"`
	Status    UserStatus `pg:"-"`
	StatusSQL string     `pg:"status,use_zero"`
	Email     string     `pg:"email,unique,notnull"`
	Name      string     `pg:"name,notnull"`
	Avatar    string     `pg:"avatar"`
	CreatedAt time.Time  `pg:"created_at,notnull,default:now()"`
	UpdatedAt time.Time  `pg:"updated_at,notnull,default:now()"`
}

// emailRegex is a basic email validation pattern.
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// Validate checks if the user has all required fields and valid data.
// Returns a ValidationError with field-specific context on failure.
//
// Validations performed:
//   - Email: required, valid format
//   - Name: required
//   - Status: valid enum value
//
// UUID and timestamps are not validated as they are managed by the repository/database.
func (u *User) Validate() error {
	if u.Email == "" {
		return newValidationError("email", "is required")
	}

	if !emailRegex.MatchString(u.Email) {
		return newValidationError("email", "invalid format")
	}

	if u.Name == "" {
		return newValidationError("name", "is required")
	}

	if u.Status < UserActive || u.Status >= userStatusUnsupported {
		return newValidationError("status", "invalid value")
	}

	return nil
}

// BeforeInsert prepares the user before insertion into the database.
// It generates a UUID if missing and converts the Status enum to its string form.
func (u *User) BeforeInsert(ctx context.Context) (context.Context, error) {
	if u.UUID == uuid.Nil {
		newUUID, err := uuid.NewV4()
		if err != nil {
			return ctx, fmt.Errorf("generate UUID: %w", err)
		}
		u.UUID = newUUID
	}

	status := u.Status.String()
	if status == "" {
		return ctx, fmt.Errorf("invalid status value: %d", u.Status)
	}

	u.StatusSQL = status

	return ctx, nil
}

// BeforeUpdate prepares the user before update in the database.
// It converts the Status enum to its string form and updates the timestamp.
func (u *User) BeforeUpdate(ctx context.Context) (context.Context, error) {
	status := u.Status.String()
	if status == "" {
		return ctx, fmt.Errorf("invalid status value: %d", u.Status)
	}

	u.StatusSQL = status
	u.UpdatedAt = time.Now()

	return ctx, nil
}

// AfterSelect processes the user after retrieval from the database.
// It converts the StatusSQL string back to the Status enum.
func (u *User) AfterSelect(_ context.Context) error {
	status, err := UserStatusFromString(u.StatusSQL)
	if err != nil {
		return fmt.Errorf("parse user status: %w", err)
	}

	u.Status = status

	return nil
}
