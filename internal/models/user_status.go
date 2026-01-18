package models

import (
	"fmt"
	"strings"
)

// UserStatus represents the current state of a user account.
type UserStatus int

const (
	// UserActive indicates an active, verified user account.
	UserActive UserStatus = iota

	// UserDeleted indicates a deleted user account.
	UserDeleted

	// userStatusUnsupported is an internal sentinel for invalid statuses.
	userStatusUnsupported
)

// userStatuses maps UserStatus enum values to their string representations.
var userStatuses = [...]string{
	UserActive:  "active",
	UserDeleted: "deleted",
}

// String returns the string representation of the UserStatus.
func (s UserStatus) String() string {
	if s < 0 || int(s) >= len(userStatuses) {
		return ""
	}
	return userStatuses[s]
}

// UserStatusFromString parses a string into a UserStatus enum.
// The comparison is case-insensitive.
func UserStatusFromString(s string) (UserStatus, error) {
	for i, r := range userStatuses {
		if strings.EqualFold(s, r) {
			return UserStatus(i), nil
		}
	}
	return userStatusUnsupported, fmt.Errorf("invalid user status value %q", s)
}

// TODO: Add proto conversion methods when integrating gRPC:
//   - Proto() proto.UserStatus
//   - UserStatusFromProto(s proto.UserStatus) UserStatus
//   - UserStatusFromProtoError(s proto.UserStatus) (UserStatus, error)
