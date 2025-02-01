package server

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Mahider-T/autoSphere/internal/database"
)

func (ser Server) catCreate(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Id    int    `json:"id"`
		Value string `json:"value"`
	}

	err := ser.readJSON(w, r, &input)
	if err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}
	cat := database.Category{
		Id:    input.Id,
		Value: input.Value,
	}
	if err = ser.models.Category.Create(&cat); err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/users/%d", cat.Id))
	err = ser.writeJSON(w, http.StatusCreated, envelope{"category": cat}, headers)
	if err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}
}
func (ser Server) catGetOne(w http.ResponseWriter, r *http.Request) {
	id, err := ser.readIDParam(r)
	if err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}
	cat, err := ser.models.Category.Get(id)
	if err != nil {
		if errors.Is(err, database.ErrRecordNotFound) {
			ser.notFoundResponse(w, r)
			return
		}
		ser.serverErrorResponse(w, r, err)
		return
	}
	ser.writeJSON(w, http.StatusOK, envelope{"category": cat}, nil)
}

func (ser Server) catGetAll(w http.ResponseWriter, r *http.Request) {
	cats, total, err := ser.models.Category.GetAll()
	if err != nil {
		if errors.Is(err, database.ErrRecordNotFound) {
			ser.notFoundResponse(w, r)
			return
		}
		ser.serverErrorResponse(w, r, err)
		return
	}
	err = ser.writeJSON(w, http.StatusOK, envelope{"total": total, "categoreis": cats}, nil)
	if err != nil {
		ser.serverErrorResponse(w, r, err)
	}
}

func (ser Server) catPut(w http.ResponseWriter, r *http.Request) {

	id, err := ser.readIDParam(r)
	if err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}

	cat, err := ser.models.Category.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrRecordNotFound):
			ser.notFoundResponse(w, r)
		default:
			ser.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		Value string `json:"value"`
	}

	err = ser.readJSON(w, r, &input)

	if err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}
	cat.Value = input.Value

	if err = ser.models.Category.Put(cat); err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}
	err = ser.writeJSON(w, http.StatusOK, envelope{"category": cat}, nil)

	if err != nil {
		ser.serverErrorResponse(w, r, err)
	}

}

func (ser Server) catDelete(w http.ResponseWriter, r *http.Request) {
	id, err := ser.readIDParam(r)
	if err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}

	_, err = ser.models.Category.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrRecordNotFound):
			ser.notFoundResponse(w, r)
		default:
			ser.serverErrorResponse(w, r, err)
			return
		}
	}

	if err = ser.models.Category.Delete(id); err != nil {
		ser.notFoundResponse(w, r)
		return
	}
}
