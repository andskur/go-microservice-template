package repository

import (
	"fmt"

	"github.com/go-pg/pg/v10/orm"

	"microservice-template/internal/models"
)

// UserGetter represents available user query strategies.
type UserGetter int

// User getter constants.
const (
	// UserUUID getter uses the user's UUID field.
	UserUUID UserGetter = iota

	// Email getter uses the user's Email field.
	Email
)

// userGetters is a slice of UserGetter string representations.
var userGetters = [...]string{
	UserUUID: "uuid",
	Email:    "email",
}

// String returns UserGetter enum as a string.
func (g UserGetter) String() string {
	return userGetters[g]
}

// Validate checks if the getter is supported.
func (g UserGetter) Validate() error {
	if g < 0 || int(g) >= len(userGetters) {
		return fmt.Errorf("unsupported user getter: %d", g)
	}
	return nil
}

// Get applies the getter to a database query.
// This is used to add WHERE clauses based on the getter type.
func (g UserGetter) Get(query *orm.Query, model *models.User) error {
	switch g {
	case UserUUID:
		query.WherePK()
	case Email:
		query.Where(fmt.Sprintf("%s = ?", g.String()), model.Email)
	default:
		return fmt.Errorf("unsupported user getter: %s", g)
	}

	return nil
}
