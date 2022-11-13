////filename: cmd/api/users.go

package main

import (
	"errors"
	"net/http"

	"grocery.jamesfaber.net/internal/data"
	"grocery.jamesfaber.net/internal/validator"
)

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	//Hold data from reuest body
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	//Parese the request body into the anonymous struct
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	//copy the data to a new struct
	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}

	//generate a password hash
	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Perform validation
	v := validator.New()
	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	//Insert the datain the database
	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exist")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	app.background(func() {
		// Send the email to the new user
		err = app.mailer.Send(user.Email, "user_welcome.tmpl", user)
		if err != nil {
			app.logger.PrintError(err, nil)
			return
		}
	})

	//write a 201 created status
	err = app.writeJSON(w, http.StatusCreated, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// // background accepts a function as its parameter
// func (app *application) background(fn func()) {
// 	//launch/create a goroutine which runs an anonymous function that sends the welcome message
// 	go func() {
// 		//recover from panics
// 		defer func() {
// 			if err := recover(); err != nil {
// 				app.logger.PrintError(fmt.Errorf("%s", err), nil)
// 			}
// 		}()
// 		// execute fn which is code to send an email
// 		fn()
// 	}()

// }
