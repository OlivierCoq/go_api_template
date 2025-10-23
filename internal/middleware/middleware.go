package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/OlivierCoq/go_api_template/internal/store"
	"github.com/OlivierCoq/go_api_template/internal/utils"
)

type UserMiddleware struct {
	UserStore store.UserStore
}

/*
	Collisions.
	The reason we store context key in a separate type is to avoid collisions with other context keys.
	If we used a string type for the context key, it could potentially collide with other context keys
	used by other packages or libraries that also use string keys.
	By using a separate type, we ensure that our context key is unique and cannot collide with other keys.
*/

type contextKey string

const userContextKey = contextKey("user")

func SetUser(r *http.Request, user *store.User) *http.Request {
	// Insert user into context property of the request. Every http request has a context property:
	// We will do this even with anonymous users, so that downstream handlers can always expect a user to be present in the context.
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

// Log off user:
func (um *UserMiddleware) Logout(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Always add the Vary header when dealing with authentication:
		w.Header().Add("Vary", "Authorization") // Caching proxies should consider the Authorization header when deciding whether to serve a cached response.
		// Remove the user from the context by setting it to anonymous user:
		r = SetUser(r, store.AnonymousUser)
		next.ServeHTTP(w, r)
	})
}

// get user from context:
func GetUser(r *http.Request) *store.User {
	// So now, we can retrieve the user from the context:
	user, ok := r.Context().Value(userContextKey).(*store.User)
	if !ok {
		panic("missing user in request") // bad actor call. This could be a hacker trying to access a protected route without authentication.
	}
	return user
}

// Middleware function to authenticate user based on Bearer token in Authorization header:
func (um *UserMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Always add the Vary header when dealing with authentication:
		// The Vary header indicates to caching proxies that the response may vary based on the value of the Authorization header.
		w.Header().Add("Vary", "Authorization") // Caching proxies should consider the Authorization header when deciding whether to serve a cached response.
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// No auth header, so we set the user as anonymous and proceed to the next handler:
			r = SetUser(r, store.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		// Check the format of the Authorization header:
		headerParts := strings.Split(authHeader, " ") // Bearer tokenstring
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			// Invalid auth header format, so we set the user as anonymous and proceed to the next handler:
			utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "invalid authorization header format"})
			return
		}

		token := headerParts[1]
		user, err := um.UserStore.GetUserToken("authentication", token)
		if err != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "failed to retrieve user"})
			return
		}
		if user == nil {
			// No user found for the provided token, so we set the user as anonymous and proceed to the next handler:
			utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "invalid or expired token"})
			return
		}

		// User found, set it in the context and proceed to the next handler:
		r = SetUser(r, user)
		next.ServeHTTP(w, r)

	})
}

// Handler function from routes to protect routes that require authentication:
func (um *UserMiddleware) RequireUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUser(r)
		if user.IsAnonymous() {
			utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "you must be authenticated to access this resource"})
			return
		}
		next.ServeHTTP(w, r)
	})
}
