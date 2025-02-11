package server

import (
	"fmt"
	"net/http"
)

func (ser *Server) logError(r *http.Request, err error) {
	// ser.logger.Println(err)
	ser.logger.PrintError(err, map[string]string{
		"request_method": r.Method,
		"request_url":    r.URL.String(),
	})
}

func (ser *Server) errorResponse(w http.ResponseWriter, r *http.Request, status int, message interface{}) {

	envelope := envelope{"error": message}

	if err := ser.writeJSON(w, http.StatusInternalServerError, envelope, nil); err != nil {
		ser.logError(r, err)
	}

}

func (ser *Server) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	ser.logError(r, err)

	message := "The server has enountered an error and could not process your request"
	ser.errorResponse(w, r, http.StatusInternalServerError, message)
}

func (ser *Server) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "The requested resource was not found"
	ser.errorResponse(w, r, http.StatusNotFound, message)
}

func (ser *Server) methodNotAllowd(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("the %s method is not allowd for this resource", r.Method)
	ser.errorResponse(w, r, http.StatusMethodNotAllowed, message)
}

func (ser *Server) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	ser.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}

func (ser *Server) authenticationRequiredResponse(w http.ResponseWriter, r *http.Request) {
	message := "you must be authenticated to access this resource"
	ser.errorResponse(w, r, http.StatusUnauthorized, message)
}
func (ser *Server) forbiddenErrorResponse(w http.ResponseWriter, r *http.Request) {
	message := "Forbidden"
	ser.errorResponse(w, r, http.StatusForbidden, message)
}

func (ser *Server) invalidCredentials(w http.ResponseWriter, r *http.Request) {
	message := "Invalid username and password"
	ser.errorResponse(w, r, http.StatusUnauthorized, message)
}

func (ser *Server) unauthorized(w http.ResponseWriter, r *http.Request) {
	message := "Unauthorized"
	ser.errorResponse(w, r, http.StatusUnauthorized, message)
}
