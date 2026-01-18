package models

import (
	"regexp"
	"time"

	"github.com/gofrs/uuid"
)

// User represents a user in the system.
// This is a pure data model without database-specific annotations.
//
// TODO: Add go-pg tags and hooks when integrating database:
//   - tableName struct{} `pg:"users,discard_unknown_columns"`
//   - StatusSQL string `pg:"status"` with Status `pg:"-"`
//   - BeforeInsert hook for UUID generation, timestamp setting, status conversion
//   - BeforeUpdate hook for timestamp updates, status conversion
//   - AfterSelect hook for status enum conversion from StatusSQL
//
// TODO: Add conversion methods when integrating:
//   - ToJWT() map[string]interface{} for JWT claims generation
//   - Proto() *proto.User for gRPC protocol buffer conversion
//   - UserFromProto(*proto.User) *User helper for inbound proto messages
//   - UsersToProto([]*User) []*proto.User for slice conversions
type User struct {
	UUID      uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	Email     string
	Name      string
	Avatar    string
	Status    UserStatus
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

	// Validate status is a known enum value
	if u.Status < UserActive || u.Status >= userStatusUnsupported {
		return newValidationError("status", "invalid value")
	}

	return nil
}
