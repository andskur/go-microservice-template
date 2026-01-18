// Package repository defines persistence interfaces and helpers.
package repository

import "microservice-template/internal/models"

// IRepository defines the storage interface for database operations.
// This interface abstracts the underlying database implementation.
type IRepository interface {
	// CreateUser creates a new user in the database.
	CreateUser(user *models.User) error

	// UserBy retrieves a user from database using the specified getter.
	UserBy(user *models.User, getter UserGetter) error
}
