package server

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Mahider-T/autoSphere/internal/database"
)

func (ser Server) catMemberCreate(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Id    int    `json:"id"`
		Value string `json:"value"`
	}

	err := ser.readJSON(w, r, &input)
	if err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}
	catMember := database.CategoryMember{
		Id:    input.Id,
		Value: input.Value,
	}
	if err = ser.models.CategoryMember.Create(&catMember); err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/category-members/%d", catMember.Id))
	err = ser.writeJSON(w, http.StatusCreated, envelope{"category_member": catMember}, headers)
	if err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}
}

func (ser Server) catMemberGetOne(w http.ResponseWriter, r *http.Request) {
	id, err := ser.readIDParam(r)
	if err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}
	catMember, err := ser.models.CategoryMember.Get(id)
	if err != nil {
		if errors.Is(err, database.ErrRecordNotFound) {
			ser.notFoundResponse(w, r)
			return
		}
		ser.serverErrorResponse(w, r, err)
		return
	}
	ser.writeJSON(w, http.StatusOK, envelope{"category_member": catMember}, nil)
}

func (ser Server) catMemberGetAll(w http.ResponseWriter, r *http.Request) {
	catMembers, total, err := ser.models.CategoryMember.GetAll()
	if err != nil {
		if errors.Is(err, database.ErrRecordNotFound) {
			ser.notFoundResponse(w, r)
			return
		}
		ser.serverErrorResponse(w, r, err)
		return
	}
	err = ser.writeJSON(w, http.StatusOK, envelope{"total": total, "category_members": catMembers}, nil)
	if err != nil {
		ser.serverErrorResponse(w, r, err)
	}
}

func (ser Server) catMemberPut(w http.ResponseWriter, r *http.Request) {
	id, err := ser.readIDParam(r)
	if err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}

	catMember, err := ser.models.CategoryMember.Get(id)
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
	catMember.Value = input.Value

	if err = ser.models.CategoryMember.Put(catMember); err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}
	err = ser.writeJSON(w, http.StatusOK, envelope{"category_member": catMember}, nil)
	if err != nil {
		ser.serverErrorResponse(w, r, err)
	}
}

func (ser Server) catMemberDelete(w http.ResponseWriter, r *http.Request) {
	id, err := ser.readIDParam(r)
	if err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}

	_, err = ser.models.CategoryMember.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrRecordNotFound):
			ser.notFoundResponse(w, r)
		default:
			ser.serverErrorResponse(w, r, err)
			return
		}
	}

	if err = ser.models.CategoryMember.Delete(id); err != nil {
		ser.notFoundResponse(w, r)
		return
	}
}
