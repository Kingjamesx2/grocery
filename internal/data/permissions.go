// Filename: internal/data/permissions.go
package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
)

// Define a slice to hold the permission codes
type Permissions []string

// Checks the slice for a specific permission code
func (p Permissions) Include(code string) bool {
	for i := range p {
		if code == p[i] {
			return true
		}
	}
	return false
}

type PermissionModel struct {
	DB *sql.DB
}

func (m PermissionModel) GetAllForUser(userID int64) (Permissions, error) {
	query := `
	     SELECT permissions.code
		 FROM permissions
		 INNER JOIN users_permissions
		 ON users_permissions.permission_id = permissions.id
		 INNER JOIN users
		 ON users_permissions.user_id = users.id
		 WHERE users.id = $1
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := m.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permisisons Permissions
	for rows.Next() {
		var permisison string
		err := rows.Scan(&permisison)
		if err != nil {
			return nil, err
		}
		permisisons = append(permisisons, permisison)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return permisisons, nil
}

func (m PermissionModel) AddForUser(userID int64, codes ...string) error {
	query := `
	      INSERT INTO users_permissions
		  SELECT $1, permissions.id FROM permissions WHERE permissions.code = ANY($2)	 
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, userID, pq.Array(codes))
	return err
}
