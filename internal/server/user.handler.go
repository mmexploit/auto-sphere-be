package server

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Mahider-T/autoSphere/internal/database"
)

func (ser *Server) userCreate(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Name         string        `json:"name"`
		Email        string        `json:"email"`
		Password     string        `json:"password"`
		Phone_Number string        `json:"phone_number"`
		Role         database.Role `json:"role"`
	}

	if err := ser.readJSON(w, r, &input); err != nil {
		ser.serverErrorRespone(w, r, err)
		return
	}
	//TODO : Add validator to validate the inserted user

	//Insert the user using the repository insert method
	user := database.User{
		Name:         input.Name,
		Email:        input.Email,
		Password:     input.Password,
		Phone_Number: input.Phone_Number,
		Role:         input.Role,
	}

	if err := ser.models.Users.Create(&user); err != nil {
		ser.serverErrorRespone(w, r, err)
		return
	}

	//return the result with the response write (write JSON)
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/users/%d", user.Id))
	err := ser.writeJSON(w, http.StatusCreated, envelope{"user": user}, headers)
	if err != nil {
		ser.serverErrorRespone(w, r, err)
		return
	}
}

func (ser Server) userGetOne(w http.ResponseWriter, r *http.Request) {
	id, err := ser.readIDParam(r)
	if err != nil {
		ser.serverErrorRespone(w, r, err)
		return
	}

	user, err := ser.models.Users.Get(id)

	if err != nil {
		switch {
		case errors.Is(err, database.ErrRecordNotFound):
			ser.notFoundResponse(w, r)
		default:
			ser.serverErrorRespone(w, r, err)
		}
		return
	}

	fmt.Print(user)
	err = ser.writeJSON(w, http.StatusOK, envelope{"user": user}, nil)

	if err != nil {
		ser.serverErrorRespone(w, r, err)
	}

}

func (ser Server) userDelete(w http.ResponseWriter, r *http.Request) {
	id, err := ser.readIDParam(r)
	if err != nil {
		ser.serverErrorRespone(w, r, err)
	}

	if err = ser.models.Users.Delete(id); err != nil {
		switch {
		case errors.Is(err, database.ErrRecordNotFound):
			ser.notFoundResponse(w, r)
		default:
			ser.serverErrorRespone(w, r, err)
			return
		}
	}
	err = ser.writeJSON(w, http.StatusOK, envelope{"message": "user successfully deleted"}, nil)
	if err != nil {
		ser.serverErrorRespone(w, r, err)
	}

}

func (ser Server) userPatch(w http.ResponseWriter, r *http.Request) {

	id, err := ser.readIDParam(r)
	if err != nil {
		ser.notFoundResponse(w, r)
		return
	}

	user, err := ser.models.Users.Get(id)

	if err != nil {
		switch {
		case errors.Is(err, database.ErrRecordNotFound):
			ser.notFoundResponse(w, r)
		default:
			ser.serverErrorRespone(w, r, err)
		}
		return
	}

	var input struct {
		Name         *string        `json:"name"`
		Email        *string        `json:"email"`
		Phone_Number *string        `json:"phone_number"`
		Role         *database.Role `json:"role"`
	}

	err = ser.readJSON(w, r, &input)

	if err != nil {
		ser.serverErrorRespone(w, r, err)
		return
	}

	if input.Name != nil {
		user.Name = *input.Name
	}
	if input.Email != nil {
		user.Email = *input.Email
	}
	if input.Phone_Number != nil {
		user.Phone_Number = *input.Phone_Number
	}
	if input.Role != nil {
		user.Role = *input.Role
	}

	//Todo : Validate the input here

	err = ser.models.Users.Patch(user)
	if err != nil {
		ser.serverErrorRespone(w, r, err)
		return
	}
	err = ser.writeJSON(w, http.StatusOK, envelope{"user": user}, nil)

	if err != nil {
		ser.serverErrorRespone(w, r, err)
	}

}
