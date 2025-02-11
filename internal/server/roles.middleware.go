package server

import (
	"context"
	"net/http"
	"strings"
)

type UserContextKey string

const (
	UserIDKey   UserContextKey = "userID"
	UserRoleKey UserContextKey = "userRole"
)

func (ser Server) RoleMiddleware(handler http.HandlerFunc, requiredRoles ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header is missing", http.StatusUnauthorized)
			return
		}

		tokenString := strings.Split(authHeader, " ")[1]
		if tokenString == "" {
			http.Error(w, "Token is missing", http.StatusUnauthorized)
			return
		}

		claims, err := ser.verifyToken(tokenString)
		if err != nil {
			ser.unauthorized(w, r)
			return
		}

		userRole := claims["role"].(string)
		if !roleInList(userRole, requiredRoles) {
			http.Error(w, "Forbidden: You do not have the required role", http.StatusForbidden)
			return
		}

		userID, ok := claims["sub"]
		if !ok {
			http.Error(w, "Invalid token: user ID not found", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		ctx = context.WithValue(ctx, UserRoleKey, userRole)
		r = r.WithContext(ctx)

		handler(w, r)
	}
}

// roleInList checks if the role exists in the list of required roles
func roleInList(role string, roles []string) bool {
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}
