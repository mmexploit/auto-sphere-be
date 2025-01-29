package server

import (
	"fmt"
	"net/http"
)

func (ser *Server) logError(r *http.Request, err error) {
	ser.logger.Println(err)
}

func (ser *Server) errorRespone(w http.ResponseWriter, r *http.Request, status int, message interface{}) {

	envelope := envelope{"error": message}

	if err := ser.writeJSON(w, http.StatusInternalServerError, envelope, nil); err != nil {
		ser.logError(r, err)
	}

}

func (ser *Server) serverErrorRespone(w http.ResponseWriter, r *http.Request, err error) {
	ser.logger.Println(r, err)

	message := "The server has enountered an error and could not process your request"
	ser.errorRespone(w, r, http.StatusInternalServerError, message)
}

func (ser *Server) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "The requested resource was not found"
	ser.errorRespone(w, r, http.StatusNotFound, message)
}

func (ser *Server) methodNotAllowd(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("the %s method is not allowd for this resource", r.Method)
	ser.errorRespone(w, r, http.StatusMethodNotAllowed, message)
}
