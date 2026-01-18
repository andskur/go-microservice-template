package repository

import (
	"fmt"
)

// UserGetter represents available user query strategies.
type UserGetter int

// User getter constants
const (
	UserUUID UserGetter = iota
	Email
)

// userGetters is slice of User Getters string representations
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
//
// Example implementation (when go-pg is added):
//   func (g UserGetter) Get(query *orm.Query, model *models.User) error {
//       switch g {
//       case UserUUID:
//           query.WherePK()
//       case Email:
//           query.Where(fmt.Sprintf("%s = ?", g.String()), model.Email)
//       default:
//           return fmt.Errorf("unsupported user getter: %s", g)
//       }
//       return nil
//   }
//
// TODO: Uncomment and implement when go-pg and models are added.
