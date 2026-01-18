package repository

import "testing"

// mockRepository implements IRepository for testing purposes.
type mockRepository struct{}

func (m *mockRepository) CreateUser(_ interface{}) error {
	return nil
}

func (m *mockRepository) UserBy(_ interface{}, _ UserGetter) error {
	return nil
}

func TestRepositoryInterface_MockImplementation(t *testing.T) {
	repo := &mockRepository{}

	if err := repo.CreateUser(map[string]interface{}{"email": "test@example.com"}); err != nil {
		t.Fatalf("CreateUser returned error: %v", err)
	}

	if err := repo.UserBy(map[string]interface{}{"email": "test@example.com"}, Email); err != nil {
		t.Fatalf("UserBy returned error: %v", err)
	}
}

func TestUserGetter_String(t *testing.T) {
	tests := []struct {
		getter UserGetter
		want   string
	}{
		{getter: UserUUID, want: "uuid"},
		{getter: Email, want: "email"},
	}

	for _, tt := range tests {
		if got := tt.getter.String(); got != tt.want {
			t.Errorf("UserGetter.String() = %v, want %v", got, tt.want)
		}
	}
}

func TestUserGetter_Validate(t *testing.T) {
	// Valid getters
	if err := UserUUID.Validate(); err != nil {
		t.Errorf("UserUUID should be valid: %v", err)
	}
	if err := Email.Validate(); err != nil {
		t.Errorf("Email should be valid: %v", err)
	}

	// Invalid getter
	invalid := UserGetter(999)
	if err := invalid.Validate(); err == nil {
		t.Error("Invalid getter should return error")
	}
}
