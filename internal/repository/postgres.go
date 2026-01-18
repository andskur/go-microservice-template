package repository

import (
	"fmt"

	"microservice-template/internal/models"
	"microservice-template/pkg/logger"
)

// PostgresRepository is a PostgreSQL-based repository implementation.
// This is a working example showing the repository pattern.
// Complete implementation will be added when database connection and models are ready.
type PostgresRepository struct {
	db interface{} // TODO: Change to *pg.DB when go-pg is added
}

// NewPostgresRepository creates a new PostgreSQL repository instance.
// The db parameter should be a database connection (e.g., *pg.DB).
//
// Example usage:
//
//	db := pg.Connect(&pg.Options{...})
//	repo := repository.NewPostgresRepository(db)
func NewPostgresRepository(db interface{}) IRepository {
	return &PostgresRepository{db: db}
}

// CreateUser creates a new user in PostgreSQL database.
//
// Example implementation (when models are ready):
//
//	_, err := r.db.(*pg.DB).Model(user).Returning("*").Insert()
//	if err != nil {
//	    return fmt.Errorf("insert user %s into db: %w", user.Email, err)
//	}
func (r *PostgresRepository) CreateUser(user *models.User) error {
	logger.Log().Info("PostgresRepository.CreateUser called")

	// TODO: Implement when models and go-pg are added
	// Example from your code:
	// if _, err := r.db.(*pg.DB).Model(user).Returning("*").Insert(); err != nil {
	//     return fmt.Errorf("insert user %s into db: %w", user.Email, err)
	// }

	return fmt.Errorf("not yet implemented: add models and database connection first")
}

// UserBy retrieves a user from PostgreSQL database using the specified getter.
//
// Example implementation (when models are ready):
//
//	query := r.db.(*pg.DB).Model(user).Column("user.*")
//	if err := getter.Get(query, user); err != nil {
//	    return fmt.Errorf("parse getter: %w", err)
//	}
//	if err := query.Select(); err != nil {
//	    return fmt.Errorf("get user from database by %s: %w", getter.String(), err)
//	}
func (r *PostgresRepository) UserBy(user *models.User, getter UserGetter) error {
	logger.Log().Infof("PostgresRepository.UserBy called with getter: %s", getter.String())

	// Validate getter
	if err := getter.Validate(); err != nil {
		return err
	}

	// TODO: Implement when models and go-pg are added
	// Example from your code:
	// query := r.db.(*pg.DB).Model(user).Column("user.*")
	// if err := getter.Get(query, user); err != nil {
	//     return fmt.Errorf("parse getter: %w", err)
	// }
	// if err := query.Select(); err != nil {
	//     return fmt.Errorf("get user from database by %s: %w", getter.String(), err)
	// }

	return fmt.Errorf("not yet implemented: add models and database connection first")
}
