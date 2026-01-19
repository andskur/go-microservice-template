package formatter

import (
	"github.com/go-openapi/strfmt"
	"github.com/gofrs/uuid"

	apimodels "microservice-template/internal/http/models"
	"microservice-template/internal/models"
)

// UserToAPI converts domain User model to API User model
func UserToAPI(user *models.User) *apimodels.User {
	email := strfmt.Email(user.Email)
	name := user.Name
	status := user.Status.String()
	createdAt := strfmt.DateTime(user.CreatedAt)
	updatedAt := strfmt.DateTime(user.UpdatedAt)

	return &apimodels.User{
		UUID:      strfmt.UUID(user.UUID.String()),
		Email:     &email,
		Name:      &name,
		Status:    &status,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}

// UserFromAPI converts API User model to domain User model
func UserFromAPI(apiUser *apimodels.User) *models.User {
	status, _ := models.UserStatusFromString(*apiUser.Status)

	return &models.User{
		UUID:   uuid.FromStringOrNil(apiUser.UUID.String()),
		Email:  string(*apiUser.Email),
		Name:   *apiUser.Name,
		Status: status,
	}
}
