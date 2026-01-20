// Package formatter converts between domain models and API models.
package formatter

import (
	"fmt"

	"github.com/go-openapi/strfmt"
	"github.com/gofrs/uuid"

	apiModels "microservice-template/internal/http/models"
	domainModels "microservice-template/internal/models"
)

// UserToAPI converts domain User model to API User model.
func UserToAPI(user *domainModels.User) *apiModels.User {
	if user == nil {
		return nil
	}

	status := user.Status.String()

	return &apiModels.User{
		UUID:      strfmt.UUID(user.UUID.String()),
		Email:     (*strfmt.Email)(&user.Email),
		Name:      &user.Name,
		Status:    &status,
		CreatedAt: strfmt.DateTime(user.CreatedAt),
		UpdatedAt: strfmt.DateTime(user.UpdatedAt),
	}
}

// UserFromAPI converts API User model to domain User model.
func UserFromAPI(apiUser *apiModels.User) (*domainModels.User, error) {
	if apiUser == nil {
		return nil, nil
	}

	user := &domainModels.User{}

	if apiUser.Email != nil {
		user.Email = string(*apiUser.Email)
	}

	if apiUser.UUID.String() != "" {
		uuidVal, err := uuid.FromString(apiUser.UUID.String())
		if err != nil {
			return &domainModels.User{}, nil
		}
		user.UUID = uuidVal
	}

	if apiUser.Status != nil {
		status, err := domainModels.UserStatusFromString(*apiUser.Status)
		if err != nil {
			return nil, fmt.Errorf("parse status: %w", err)
		}
		user.Status = status
	}

	if apiUser.Name != nil {
		user.Name = *apiUser.Name
	}

	return user, nil
}
