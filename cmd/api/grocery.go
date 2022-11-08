//FIlename: cmd/api/grocery.go

package main

import (
	"errors"
	"fmt"
	"net/http"

	"grocery.jamesfaber.net/internal/data"
	"grocery.jamesfaber.net/internal/validator"
)

// createGroceryInfoHandler for the "POST" /v1/groceryInfo" endpoint
func (app *application) createGroceryInfoHandler(w http.ResponseWriter, r *http.Request) {
	// Our Target decode destination
	var input struct {
		Name     string `json:"name"`
		Item     string `json:"item"`
		Location string `json:"location"`
		Price    string `json:"price"`
		Address  string `json:"address"`
		Phone    string `json:"phone"`
		Contact  string `json:"contact"`
		Email    string `json:"email"`
		Website  string `json:"website"`
	}
	// Initialize a new json.Decoder instance
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	//Copy the values from the input struct to a new grocery struct
	grocery := &data.Grocery{
		Name:     input.Name,
		Item:     input.Item,
		Location: input.Location,
		Price:    input.Price,
		Address:  input.Address,
		Phone:    input.Phone,
		Contact:  input.Contact,
		Email:    input.Email,
		Website:  input.Website,
	}
	// initialize a new Validator instance
	v := validator.New()

	//Check the map to determine if there were any validation errors
	if data.ValidateGrocery(v, grocery); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Create a Grocery Object
	err = app.models.Groceries.Insert(grocery)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	// Create a location header for the newly created resource/Grocery object
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/groceryInfo/%d", grocery.ID))
	// Write the JSON response with 201 - created status code with the body
	// being the actual grocery data and the header being the headers map
	err = app.writeJSON(w, http.StatusCreated, envelope{"grocery": grocery}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// showGroceryInfoHandlerfor the "GET" /v1/groceryinfo/:id" endpoint
func (app *application) showGroceryInfoHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	// Fetch the specific grocery task
	grocery, err := app.models.Groceries.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// Write the response by Get()
	err = app.writeJSON(w, http.StatusOK, envelope{"grocery": grocery}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateGroceryInfoHandler(w http.ResponseWriter, r *http.Request) {
	// This method does a partial replacement
	// Get the id for the grocery task that needs updating
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	// Fetch the original record from the database
	grocery, err := app.models.Groceries.Get(id)
	// Error handling
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// Create an input struct to hold data read in from the client
	// We update the input struct to use pointers because pointers have a
	// default value of nil false
	// if a field remains nil then we know that the client did not update it
	var input struct {
		Name     *string `json:"name"`
		Item     *string `json:"item"`
		Location *string `json:"location"`
		Price    *string `json:"price"`
		Address  *string `json:"address"`
		Phone    *string `json:"phone"`
		Contact  *string `json:"contact"`
		Email    *string `json:"email"`
		Website  *string `json:"website"`
	}

	//Initalize a new json.Decoder instance
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	// Check for updates
	if input.Name != nil {
		grocery.Name = *input.Name
	}
	if input.Item != nil {
		grocery.Item = *input.Item
	}
	if input.Location != nil {
		grocery.Location = *input.Location
	}
	if input.Price != nil {
		grocery.Price = *input.Price
	}
	if input.Address != nil {
		grocery.Address = *input.Address
	}
	if input.Phone != nil {
		grocery.Phone = *input.Phone
	}
	if input.Contact != nil {
		grocery.Contact = *input.Contact
	}
	if input.Email != nil {
		grocery.Email = *input.Email
	}
	if input.Website != nil {
		grocery.Website = *input.Website
	}

	// Perform Validation on the updated grocery task. If validation fails then
	// we send a 422 - unprocessable entity response to the client
	// initialize a new Validator instance
	v := validator.New()

	//Check the map to determine if there were any validation errors
	if data.ValidateGrocery(v, grocery); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// Pass the update grocery record to the Update() method
	err = app.models.Groceries.Update(grocery)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	err = app.writeJSON(w, http.StatusCreated, envelope{"grocery": grocery}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

// The deleteGroceryInfoHandler() allows the user to delete a grocery info from the databse by using the ID
func (app *application) deleteGroceryInfoHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	// Delete the grocery tasks from the database. Send a 404 Not Found status code to the
	// client if there is no matching record
	err = app.models.Groceries.Delete(id)
	// Error handling
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// Return 200 Status OK to the client with a success message
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "grocery info successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// The listGroceryInfoHandler() allows the client to see a listing of grocery tasks
// based on a set criteria
func (app *application) listGroceryInfoHandler(w http.ResponseWriter, r *http.Request) {
	// Create an input struct to hold our query parameter
	var input struct {
		Name     string
		Item     string
		Location string
		Price    string
		Address  string
		Phone    string
		Contact  string
		Email    string
		Website  string
		data.Filters
	}
	// Initialize a validator
	v := validator.New()
	// Get the URL values map
	qs := r.URL.Query()
	// use the helper methods to extract values
	input.Name = app.readString(qs, "name", "")
	input.Item = app.readString(qs, "item", "")
	input.Location = app.readString(qs, "location", "")
	input.Price = app.readString(qs, "price", "")
	input.Address = app.readString(qs, "address", "")
	input.Phone = app.readString(qs, "phone", "")
	input.Contact = app.readString(qs, "contact", "")
	input.Email = app.readString(qs, "email", "")
	input.Website = app.readString(qs, "website", "")
	//------------------------------------------------------------------------
	// Get the page information using the read int method
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	// Get the sort information
	input.Filters.Sort = app.readString(qs, "sort", "id")
	// Specify the allowed sort values
	input.Filters.SortList = []string{"id", "name", "item", "location", "price", "address", "phone", "contact", "email", "website",
		"-id", "-name", "-item", "-location", "price", "address", "-phone", "-contact", "-email", "-website"}
	// Check for validation errors
	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// Get a listing of all grocery tasks
	groceries, metadata, err := app.models.Groceries.GetAll(input.Name, input.Item, input.Price, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// Send a JSON response containing all the grocery tasks
	err = app.writeJSON(w, http.StatusOK, envelope{"groceries": groceries, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}
