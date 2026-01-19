package repository

import (
	"fmt"

	"github.com/go-pg/pg/v10"

	"microservice-template/internal/models"
	"microservice-template/pkg/logger"
)

// PostgresRepository is a PostgreSQL-based repository implementation using go-pg.
type PostgresRepository struct {
	db *pg.DB
}

// NewPostgresRepository creates a new PostgreSQL repository instance.
// The db parameter must be a connected *pg.DB instance.
func NewPostgresRepository(db *pg.DB) IRepository {
	return &PostgresRepository{db: db}
}

// CreateUser creates a new user in PostgreSQL database.
// The user's UUID will be auto-generated if not set (via BeforeInsert hook).
// Returns the created user with all fields populated.
func (r *PostgresRepository) CreateUser(user *models.User) error {
	logger.Log().Infof("creating user: %s", user.Email)

	if _, err := r.db.Model(user).Returning("*").Insert(); err != nil {
		return fmt.Errorf("insert user %s into db: %w", user.Email, err)
	}

	logger.Log().Infof("user created successfully: %s (UUID: %s)", user.Email, user.UUID)
	return nil
}

// UserBy retrieves a user from PostgreSQL database using the specified getter.
// The user parameter should have the field(s) required by the getter pre-populated.
func (r *PostgresRepository) UserBy(user *models.User, getter UserGetter) error {
	logger.Log().Infof("fetching user by %s", getter.String())

	if err := getter.Validate(); err != nil {
		return err
	}

	query := r.db.Model(user).Column("user.*")

	if err := getter.Get(query, user); err != nil {
		return fmt.Errorf("parse getter: %w", err)
	}

	if err := query.Select(); err != nil {
		return fmt.Errorf("get user from database by %s: %w", getter.String(), err)
	}

	logger.Log().Infof("user fetched successfully: %s (UUID: %s)", user.Email, user.UUID)
	return nil
}
