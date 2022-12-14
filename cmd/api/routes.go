//FILename: cmd/api/routes

package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	//Create a new  httprouter ruter instance
	router := httprouter.New()
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodGet, "/v1/groceryInfo", app.requirePermission("grocery:read", app.listGroceryInfoHandler))

	router.HandlerFunc(http.MethodPost, "/v1/groceryInfo", app.requirePermission("grocery:write", app.createGroceryInfoHandler))
	router.HandlerFunc(http.MethodGet, "/v1/groceryInfo/:id", app.requirePermission("grocery:read", app.showGroceryInfoHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/groceryInfo/:id", app.requirePermission("grocery:write", app.updateGroceryInfoHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/groceryInfo/:id", app.requirePermission("grocery:write", app.deleteGroceryInfoHandler))

	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	return app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(router))))
}
