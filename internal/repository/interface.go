package repository

// IRepository defines the storage interface for database operations.
// This interface abstracts the underlying database implementation.
type IRepository interface {
	// CreateUser creates a new user in the database.
	CreateUser(model interface{}) error

	// UserBy retrieves a user from database using the specified getter.
	UserBy(model interface{}, getter UserGetter) error
}
