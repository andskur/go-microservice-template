package models

import "testing"

func TestUserStatus_String(t *testing.T) {
	tests := []struct {
		status UserStatus
		want   string
	}{
		{status: UserActive, want: "active"},
		{status: UserDeleted, want: "deleted"},
		{status: userStatusUnsupported, want: ""},
		{status: UserStatus(-1), want: ""},
		{status: UserStatus(999), want: ""},
	}

	for _, tt := range tests {
		if got := tt.status.String(); got != tt.want {
			t.Fatalf("String() = %q, want %q", got, tt.want)
		}
	}
}

func TestUserStatusFromString_Valid(t *testing.T) {
	tests := map[string]UserStatus{
		"active":  UserActive,
		"deleted": UserDeleted,
		"Active":  UserActive,
		"DELETED": UserDeleted,
	}

	for input, want := range tests {
		got, err := UserStatusFromString(input)
		if err != nil {
			t.Fatalf("UserStatusFromString(%q) returned error: %v", input, err)
		}
		if got != want {
			t.Fatalf("UserStatusFromString(%q) = %v, want %v", input, got, want)
		}
	}
}

func TestUserStatusFromString_Invalid(t *testing.T) {
	if _, err := UserStatusFromString("unknown"); err == nil {
		t.Fatal("expected error for invalid status, got nil")
	}
}

func TestUserStatus_RoundTrip(t *testing.T) {
	statuses := []UserStatus{UserActive, UserDeleted}
	for _, s := range statuses {
		str := s.String()
		parsed, err := UserStatusFromString(str)
		if err != nil {
			t.Fatalf("round-trip parse failed for %q: %v", str, err)
		}
		if parsed != s {
			t.Fatalf("round-trip mismatch: got %v, want %v", parsed, s)
		}
	}
}
