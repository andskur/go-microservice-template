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

	// TODO: Additional repository methods (uncomment and implement as needed):
	//
	// // GetOrCreateUser retrieves a user by getter or creates if not found.
	// // Returns (created bool, error). If created=true, user was inserted.
	// GetOrCreateUser(user *models.User, getter UserGetter) (bool, error)
	//
	// // UpdateUserBy updates a user matching the getter with specified columns.
	// // Pass column names to update selectively, or no columns to update all.
	// UpdateUserBy(user *models.User, getter UserGetter, columns ...string) error
	//
	// // AllUsers retrieves all users from the database.
	// // Use with caution in production - consider pagination for large datasets.
	// AllUsers() ([]*models.User, error)
}
