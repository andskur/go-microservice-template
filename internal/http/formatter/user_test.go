package formatter

import (
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/gofrs/uuid"

	apimodels "microservice-template/internal/http/models"
	"microservice-template/internal/models"
)

func TestUserToAPI(t *testing.T) {
	// Create a domain user
	userUUID := uuid.Must(uuid.NewV4())
	createdAt := time.Now().UTC()
	updatedAt := createdAt.Add(time.Hour)

	domainUser := &models.User{
		UUID:      userUUID,
		Email:     "test@example.com",
		Name:      "Test User",
		Status:    models.UserActive,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	// Convert to API model
	apiUser := UserToAPI(domainUser)

	// Verify conversion
	if apiUser == nil {
		t.Fatal("UserToAPI returned nil")
	}

	if string(apiUser.UUID) != userUUID.String() {
		t.Errorf("UUID mismatch: expected %s, got %s", userUUID.String(), apiUser.UUID)
	}

	if apiUser.Email == nil {
		t.Fatal("Email is nil")
	}
	if string(*apiUser.Email) != "test@example.com" {
		t.Errorf("Email mismatch: expected test@example.com, got %s", *apiUser.Email)
	}

	if apiUser.Name == nil {
		t.Fatal("Name is nil")
	}
	if *apiUser.Name != "Test User" {
		t.Errorf("Name mismatch: expected 'Test User', got '%s'", *apiUser.Name)
	}

	if apiUser.Status == nil {
		t.Fatal("Status is nil")
	}
	if *apiUser.Status != "active" {
		t.Errorf("Status mismatch: expected 'active', got '%s'", *apiUser.Status)
	}

	// Check timestamps
	expectedCreatedAt := strfmt.DateTime(createdAt)
	if apiUser.CreatedAt != expectedCreatedAt {
		t.Errorf("CreatedAt mismatch")
	}

	expectedUpdatedAt := strfmt.DateTime(updatedAt)
	if apiUser.UpdatedAt != expectedUpdatedAt {
		t.Errorf("UpdatedAt mismatch")
	}
}

func TestUserToAPI_WithDeletedStatus(t *testing.T) {
	domainUser := &models.User{
		UUID:      uuid.Must(uuid.NewV4()),
		Email:     "deleted@example.com",
		Name:      "Deleted User",
		Status:    models.UserDeleted,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	apiUser := UserToAPI(domainUser)

	if apiUser == nil {
		t.Fatal("UserToAPI returned nil")
	}

	if apiUser.Status == nil {
		t.Fatal("Status is nil")
	}
	if *apiUser.Status != "deleted" {
		t.Errorf("Status mismatch: expected 'deleted', got '%s'", *apiUser.Status)
	}
}

func TestUserFromAPI(t *testing.T) {
	// Create an API user
	userUUID := uuid.Must(uuid.NewV4())
	email := strfmt.Email("api@example.com")
	name := "API User"
	status := "active"

	apiUser := &apimodels.User{
		UUID:   strfmt.UUID(userUUID.String()),
		Email:  &email,
		Name:   &name,
		Status: &status,
	}

	// Convert to domain model
	domainUser, err := UserFromAPI(apiUser)
	if err != nil {
		t.Fatalf("UserFromAPI returned error: %v", err)
	}

	// Verify conversion
	if domainUser == nil {
		t.Fatal("UserFromAPI returned nil")
	}

	if domainUser.UUID.String() != userUUID.String() {
		t.Errorf("UUID mismatch: expected %s, got %s", userUUID.String(), domainUser.UUID.String())
	}

	if domainUser.Email != "api@example.com" {
		t.Errorf("Email mismatch: expected api@example.com, got %s", domainUser.Email)
	}

	if domainUser.Name != "API User" {
		t.Errorf("Name mismatch: expected 'API User', got '%s'", domainUser.Name)
	}

	if domainUser.Status != models.UserActive {
		t.Errorf("Status mismatch: expected UserActive, got %v", domainUser.Status)
	}
}

func TestUserFromAPI_WithDeletedStatus(t *testing.T) {
	email := strfmt.Email("deleted@example.com")
	name := "Deleted User"
	status := "deleted"

	apiUser := &apimodels.User{
		UUID:   strfmt.UUID(uuid.Must(uuid.NewV4()).String()),
		Email:  &email,
		Name:   &name,
		Status: &status,
	}

	domainUser, err := UserFromAPI(apiUser)
	if err != nil {
		t.Fatalf("UserFromAPI returned error: %v", err)
	}

	if domainUser == nil {
		t.Fatal("UserFromAPI returned nil")
	}

	if domainUser.Status != models.UserDeleted {
		t.Errorf("Status mismatch: expected UserDeleted, got %v", domainUser.Status)
	}
}

func TestUserFromAPI_WithInvalidUUID(t *testing.T) {
	email := strfmt.Email("test@example.com")
	name := "Test User"
	status := "active"

	apiUser := &apimodels.User{
		UUID:   strfmt.UUID("invalid-uuid"),
		Email:  &email,
		Name:   &name,
		Status: &status,
	}

	domainUser, err := UserFromAPI(apiUser)
	if err != nil {
		t.Fatalf("UserFromAPI returned error: %v", err)
	}

	if domainUser == nil {
		t.Fatal("UserFromAPI returned nil")
	}

	// FromStringOrNil should return nil UUID for invalid input
	if domainUser.UUID != uuid.Nil {
		t.Errorf("Expected nil UUID for invalid input, got %s", domainUser.UUID)
	}
}

func TestUserRoundTrip(t *testing.T) {
	// Create original domain user
	original := &models.User{
		UUID:      uuid.Must(uuid.NewV4()),
		Email:     "roundtrip@example.com",
		Name:      "Round Trip User",
		Status:    models.UserActive,
		CreatedAt: time.Now().UTC().Truncate(time.Millisecond),
		UpdatedAt: time.Now().UTC().Truncate(time.Millisecond),
	}

	// Convert to API and back to domain
	apiUser := UserToAPI(original)
	converted, err := UserFromAPI(apiUser)
	if err != nil {
		t.Fatalf("UserFromAPI returned error: %v", err)
	}

	// Verify round trip (note: timestamps are not included in FromAPI)
	if converted.UUID.String() != original.UUID.String() {
		t.Errorf("UUID mismatch after round trip")
	}

	if converted.Email != original.Email {
		t.Errorf("Email mismatch after round trip")
	}

	if converted.Name != original.Name {
		t.Errorf("Name mismatch after round trip")
	}

	if converted.Status != original.Status {
		t.Errorf("Status mismatch after round trip")
	}
}
