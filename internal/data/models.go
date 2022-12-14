// Filename: internal/data/models.go
package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

// A wrapper for our data models
type Models struct {
	Permissions PermissionModel
	Groceries   GroceryModel
	Users       UserModel
	Tokens      TokenModel
}

// NewModels() allows us to create a new model
func NewModels(db *sql.DB) Models {
	return Models{
		Permissions: PermissionModel{DB: db},
		Groceries:   GroceryModel{DB: db},
		Users:       UserModel{DB: db},
		Tokens:      TokenModel{DB: db},
	}
}
