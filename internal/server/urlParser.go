package server

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/Mahider-T/autoSphere/internal/database"
	"github.com/Mahider-T/autoSphere/validator"
)

func (ser *Server) parseInt(r *http.Request, key string, defaultValue int, v *validator.Validator) int {
	s := r.URL.Query().Get(key)

	if s == "" {
		return defaultValue
	}
	val, err := strconv.Atoi(s)

	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultValue
	}

	return val
}

func (ser *Server) parseCSV(r *http.Request, key string, defaultValue []string) ([]string, error) {
	csv := r.URL.Query().Get(key)

	if csv == "" {
		return []string{}, nil
	}

	return strings.Split(csv, ","), nil
}

func (ser *Server) parseString(r *http.Request, key string, defaultValue string) string {
	qs := r.URL.Query()
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	return s
}
func (ser *Server) parseRole(r *http.Request) *database.Role {
	qs := r.URL.Query()
	roleStr := qs.Get("role")

	if roleStr == "" {
		return nil
	}

	switch roleStr {
	case string(database.ADMIN):
		role := database.ADMIN
		return &role
	case string(database.OPERATOR):
		role := database.OPERATOR
		return &role
	case string(database.SALES):
		role := database.SALES
		return &role
	default:
		return nil
	}
}
