//filename: internal/data/grocery.go

package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"grocery.jamesfaber.net/internal/validator"
)

type Grocery struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	Name      string    `json:"name"`
	Item      string    `json:"item"`
	Location  string    `json:"location"`
	Price     string    `json:"price"`
	Address   string    `json:"address"`
	Phone     string    `json:"phone"`
	Contact   string    `json:"contact"`
	Email     string    `json:"email"`
	Website   string    `json:"website"`
	Version   int32     `json:"version"`
}

func ValidateGrocery(v *validator.Validator, grocery *Grocery) {
	// Use the check() method to execute our validation checks
	v.Check(grocery.Name != "", "name", "must be provided")
	v.Check(len(grocery.Name) <= 200, "name", "must not be more than 200 bytes long")

	v.Check(grocery.Item != "", "item", "must be provided")
	v.Check(len(grocery.Item) <= 200, "item", "must not be more than 200 bytes long")

	v.Check(grocery.Location != "", "location", "must be provided")
	v.Check(len(grocery.Location) <= 200, "location", "must not be more than 200 bytes long")

	v.Check(grocery.Price != "", "price", "must be provided")
	v.Check(len(grocery.Price) <= 200, "price", "must not be more than 200 bytes long")

	v.Check(grocery.Address != "", "address", "must be provided")
	v.Check(len(grocery.Address) <= 200, "address", "must not be more than 200 bytes long")

	v.Check(grocery.Phone != "", "phone", "must be provided")
	v.Check(len(grocery.Phone) <= 200, "phone", "must not be more than 200 bytes long")

	v.Check(grocery.Contact != "", "contact", "must be provided")
	v.Check(len(grocery.Contact) <= 200, "contact", "must not be more than 200 bytes long")

	v.Check(grocery.Email != "", "email", "must be provided")
	v.Check(len(grocery.Email) <= 200, "email", "must not be more than 200 bytes long")

	v.Check(grocery.Website != "", "website", "must be provided")
	v.Check(len(grocery.Website) <= 200, "website", "must not be more than 200 bytes long")
}

// Define a grocery list model which wraps a sql.DB connection pool
type GroceryModel struct {
	DB *sql.DB
}

// Insert() allows us to create a new grocery item
func (m GroceryModel) Insert(grocery *Grocery) error {
	query := `
	INSERT INTO grocery (name, item, location, price, address, phone, contact, email, website )
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	RETURNING id, created_at, version
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// Cleanup to prevent memory leaks
	defer cancel()

	// Collect the data fields into a slice
	args := []interface{}{grocery.Name, grocery.Item, grocery.Location, grocery.Price, grocery.Address, grocery.Phone, grocery.Contact, grocery.Email, grocery.Website}
	return m.DB.QueryRowContext(ctx, query, args...).Scan(&grocery.ID, &grocery.CreatedAt, &grocery.Version)
}

// GET() allows us to retrieve a specific grocery item
func (m GroceryModel) Get(id int64) (*Grocery, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	// Create query
	query := `
		SELECT id, created_at, name, item, location, price, address, phone, contact, email, website, version
		FROM grocery
		WHERE id = $1
	`
	// Declare a Grocery variable to hold the return data
	var grocery Grocery
	// Execute Query using the QueryRow
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// Cleanup to prevent memory leaks
	defer cancel()
	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&grocery.ID,
		&grocery.CreatedAt,
		&grocery.Name,
		&grocery.Item,
		&grocery.Location,
		&grocery.Price,
		&grocery.Address,
		&grocery.Phone,
		&grocery.Contact,
		&grocery.Email,
		&grocery.Website,
		&grocery.Version,
	)
	// Handle any errors
	if err != nil {
		// Check the type of error
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	// Success
	return &grocery, nil
}

// Update() allows us to edit/alter a grocery item in the list
func (m GroceryModel) Update(grocery *Grocery) error {
	query := `
		UPDATE grocery 
		set name = $1, item = $2, location = $3, price = $4, address = $5, phone = $6, contact = $7, email = $8, website = $9
		version = version + 1
		WHERE id = $10
		AND version = $11
		RETURNING version
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// Cleanup to prevent memory leaks
	defer cancel()

	args := []interface{}{
		grocery.Name,
		grocery.Item,
		grocery.Location,
		grocery.Price,
		grocery.Address,
		grocery.Phone,
		grocery.Contact,
		grocery.Email,
		grocery.Website,
		grocery.ID,
		grocery.Version,
	}
	// Check for edit conflicts
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&grocery.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

// Delete() removes a specific item
func (m GroceryModel) Delete(id int64) error {
	// Ensure that there is a valid id
	if id < 1 {
		return ErrRecordNotFound
	}
	// Create the delete query
	query := `
		DELETE FROM grocery
		WHERE id = $1
	`
	// Create a context
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// Cleanup to prevent memory leaks
	defer cancel()

	// Execute the query
	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	// Check how many rows were affected by the delete operation. We
	// call the RowsAffected() method on the result variable
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	// Check if no rows were affected
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

// the GetAll() method returns a list of all the Grocery sorted by id
func (m GroceryModel) GetAll(name string, item string, filters Filters) ([]*Grocery, Metadata, error) {
	//construct the query to return all grocery
	//make query into formated string to be able to sort by field and asc or dec dynaimicaly
	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(),id, created_at, name, item, 
		location, price, address, phone, contact, email, website, version
		FROM grocery
		WHERE (to_tsvector('simple',name) @@ plainto_tsquery('simple', $1) OR $1 = '')
		AND (to_tsvector('simple',item) @@ plainto_tsquery('simple', $2) OR $2 = '')
		ORDER BY %s %s, id ASC
		LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortOrder())

	//create a 3 second timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	//execute the query
	args := []interface{}{name, item, filters.limit(), filters.offset()}
	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	//close the result set
	defer rows.Close()
	//store total records
	totalRecords := 0
	//intialize an empty slice to hold the Grocery data
	groceries := []*Grocery{}
	//iterate over the rows in the result set
	for rows.Next() {
		var grocery Grocery
		//scan the values from the row into the Grocery struct
		err := rows.Scan(
			&totalRecords,
			&grocery.ID,
			&grocery.CreatedAt,
			&grocery.Name,
			&grocery.Item,
			&grocery.Location,
			&grocery.Price,
			&grocery.Address,
			&grocery.Phone,
			&grocery.Contact,
			&grocery.Email,
			&grocery.Website,
			&grocery.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		//add the grocery to our slice
		groceries = append(groceries, &grocery)
	}
	//check if any errors occured while proccessing the result set
	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}
	metadata := calculateMetaData(totalRecords, filters.Page, filters.PageSize)
	//return the result set. the slice of groceries
	return groceries, metadata, nil
}
