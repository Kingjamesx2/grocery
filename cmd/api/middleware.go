// Filename: cmd/api/middleware.go

package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"grocery.jamesfaber.net/internal/data"

	"golang.org/x/time/rate"
	"grocery.jamesfaber.net/internal/validator"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (app *application) rateLimit(next http.Handler) http.Handler {
	// Create a client type
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}
	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)
	// Launch a backaground Goroutine that removes old entries
	// from the clients map once every minute
	go func() {
		for {
			time.Sleep(time.Minute)
			// Lock before starting to cleanup
			mu.Lock()
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.config.limiter.enabled {
			// Get the IP address of the request
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}
			// Lock()
			mu.Lock()
			// Check if the IP address is in the map
			if _, found := clients[ip]; !found {
				clients[ip] = &client{limiter: rate.NewLimiter(
					rate.Limit(app.config.limiter.rps),
					app.config.limiter.burst,
				)}
			}
			// Update the last seen time of the client
			clients[ip].lastSeen = time.Now()
			// Check if request allowed
			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				app.rateLimitExceededResponse(w, r)
				return
			}

			mu.Unlock()
		} // end of enabled conditional
		next.ServeHTTP(w, r)
	})
}

// Authentication
func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add a "Vary: Authorization" header to the repsonse
		// A note to caches that the reponse may vary
		w.Header().Add("Vary", "Authorization")
		// Retrieve the value of the Authorization header from the request
		authorizationHeader := r.Header.Get("Authorization")
		// If no authorization found, then we will create an anonymous user
		if authorizationHeader == "" {
			r = app.contextSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}
		// Check if the provided Authorizaation header is in the right format
		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}
		// Extract the token
		token := headerParts[1]
		// Validate the token
		v := validator.New()
		if data.ValidateTokenPlainText(v, token); !v.Valid() {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}
		// Retrieve detials about the user
		user, err := app.models.Users.GetForToken(data.ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)

			}
			return
		}
		// Add the user information to the request context
		r = app.contextSetUser(r, user)
		// Call the next handler in the chain
		next.ServeHTTP(w, r)
	})
}

// Check for activated user
func (app *application) requireAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the user
		user := app.contextGetUser(r)
		// Check for anonymous user
		if user.IsAnonymous() {
			app.authenticationRequiredResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (app *application) requireActivatedUser(next http.HandlerFunc) http.HandlerFunc {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the user
		user := app.contextGetUser(r)
		// Check if user is activated
		if !user.Activated {
			app.inactiveAccountResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
	return app.requireAuthenticatedUser(fn)
}

// Check for user permission
func (app *application) requirePermission(code string, next http.HandlerFunc) http.HandlerFunc {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the user
		user := app.contextGetUser(r)
		// Get the permission slice for the user
		permissions, err := app.models.Permissions.GetAllForUser(user.ID)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		// Check for the permission
		if !permissions.Include(code) {
			app.notPermittedResponse(w, r)
			return
		}
		// OK
		next.ServeHTTP(w, r)
	})
	return app.requireActivatedUser(fn)
}

// Enable CORS
func (app *application) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//Add the "Vary:Origin" headers
		w.Header().Add("Vary", "Origin")
		//Get the value of the request's origin's headers
		origin := r.Header.Get("Origin")
		//check if origin header present
		if origin != "" {
			for i := range app.config.cors.trustedOrigins {
				if origin == app.config.cors.trustedOrigins[i] {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					break
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}
