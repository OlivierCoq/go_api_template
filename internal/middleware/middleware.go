package middleware

import (
	"context"
	"net/http"

	"github.com/OlivierCoq/go_api_template/internal/store"
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

func GetUser(r *http.Request) *store.User {
	// So now, we can retrieve the user from the context:
	user, ok := r.Context().Value(userContextKey).(*store.User)
	if !ok {
		panic("missing user in request") // bad actor call. This could be a hacker trying to access a protected route without authentication.
	}
	return user
}
