package server

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Mahider-T/autoSphere/internal/database"
	"github.com/Mahider-T/autoSphere/validator"
)

func (ser *Server) urlParameter(w http.ResponseWriter, r *http.Request) {
	fmt.Print(r.URL.Query())
}

func (ser *Server) userCreate(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Name         string        `json:"name"`
		Email        string        `json:"email"`
		Password     string        `json:"password"`
		Phone_Number string        `json:"phone_number"`
		Role         database.Role `json:"role"`
	}

	if err := ser.readJSON(w, r, &input); err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}
	//TODO : Add validator to validate the inserted user

	user := database.User{
		Name:  input.Name,
		Email: input.Email,
		// Password:     input.Password,
		Phone_Number: input.Phone_Number,
		Role:         input.Role,
	}
	v := validator.New()
	if database.ValidateUser(v, &user); !v.Valid() {
		ser.failedValidationResponse(w, r, v.Errors)
		return
	}

	err := user.Password.Set(input.Password)
	if err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}

	if err := ser.models.Users.Create(&user); err != nil {
		switch {
		case errors.Is(err, database.ErrDuplicateEmail):
			v.AddError("email", "a user with this email already exists.")
			ser.failedValidationResponse(w, r, v.Errors)
		case errors.Is(err, database.ErrDuplicatePhoneNumber):
			v.AddError("phone number", "a user with this phone number already exists.")
			ser.failedValidationResponse(w, r, v.Errors)
		default:
			ser.serverErrorResponse(w, r, err)
		}
		return

	}

	token, err := ser.models.Tokens.New(user.Id, 3*24*time.Hour, database.ScopeActivation)
	if err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}

	go func() {

		data := map[string]interface{}{
			"userId":          user.Id,
			"activationToken": token.Plaintext,
		}
		err = ser.mailer.Send(user.Email, "user_welcome.tmpl", data)
		if err != nil {
			ser.logger.PrintError(err, nil)
		}
	}()

	//return the result with the response write (write JSON)
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/users/%d", user.Id))
	err = ser.writeJSON(w, http.StatusCreated, envelope{"user": user}, headers)
	if err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}
}

func (ser Server) userGetOne(w http.ResponseWriter, r *http.Request) {

	id, err := ser.readIDParam(r)
	if err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}

	user, err := ser.models.Users.Get(id)

	if err != nil {
		switch {
		case errors.Is(err, database.ErrRecordNotFound):
			ser.notFoundResponse(w, r)
		default:
			ser.serverErrorResponse(w, r, err)
		}
		return
	}
	err = ser.writeJSON(w, http.StatusOK, envelope{"user": user}, nil)

	if err != nil {
		ser.serverErrorResponse(w, r, err)
	}

}

func (ser Server) userDelete(w http.ResponseWriter, r *http.Request) {
	id, err := ser.readIDParam(r)
	if err != nil {
		ser.serverErrorResponse(w, r, err)
	}

	if err = ser.models.Users.Delete(id); err != nil {
		switch {
		case errors.Is(err, database.ErrRecordNotFound):
			ser.notFoundResponse(w, r)
		default:
			ser.serverErrorResponse(w, r, err)
			return
		}
	}
	err = ser.writeJSON(w, http.StatusOK, envelope{"message": "user successfully deleted"}, nil)
	if err != nil {
		ser.serverErrorResponse(w, r, err)
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
			ser.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		Name          *string        `json:"name"`
		Email         *string        `json:"email"`
		Phone_Number  *string        `json:"phone_number"`
		Role          *database.Role `json:"role"`
		Refresh_Token *string        `json:"refresh_token"`
	}

	err = ser.readJSON(w, r, &input)

	if err != nil {
		ser.serverErrorResponse(w, r, err)
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
	if input.Refresh_Token != nil {
		user.Refresh_Token = input.Refresh_Token
	}

	//Todo : Validate the input here

	err = ser.models.Users.Patch(user)
	if err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}
	err = ser.writeJSON(w, http.StatusOK, envelope{"user": user}, nil)

	if err != nil {
		ser.serverErrorResponse(w, r, err)
	}

}

func (ser Server) getUsers(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Name    string
		Role    string //should be database.Role tho
		Filters database.Filters
	}

	v := validator.New()

	//TODO : Insert validation here
	//especially role that it is one of the three valeus
	input.Name = ser.parseString(r, "name", "")

	role := ser.parseString(r, "role", "") // Capture the return value first
	input.Role = role

	var page int
	page = ser.parseInt(r, "page", 1, v)

	input.Filters.Page = page

	var page_size int
	page_size = ser.parseInt(r, "page_size", 10, v)
	input.Filters.PageSize = page_size

	input.Filters.Page = ser.parseInt(r, "page", 1, v)
	input.Filters.PageSize = ser.parseInt(r, "page_size", 20, v)
	input.Filters.Sort = ser.parseString(r, "sort", "id")

	input.Filters.SortSafelist = []string{"id", "name", "created_at", "-id", "-name", "-created_at"}

	if database.ValidateFilters(v, input.Filters); !v.Valid() {
		ser.failedValidationResponse(w, r, v.Errors)
		return
	}

	users, metadata, err := ser.models.Users.GetAll(input.Name, input.Role, input.Filters)
	if err != nil {
		ser.serverErrorResponse(w, r, err)
	}
	ser.writeJSON(w, http.StatusOK, envelope{"metadata": metadata, "users": users}, nil)

	// fmt.Fprintf(w, "%+v\n", input)
}

func (ser Server) login(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := ser.readJSON(w, r, &input); err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}

	user, err := ser.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch err {
		case database.ErrRecordNotFound:
			ser.notFoundResponse(w, r)
			return
		default:
			ser.serverErrorResponse(w, r, err)
			return
		}
	}

	matches, err := user.Password.Matches(input.Password)

	if err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}
	if !matches {
		ser.invalidCredentials(w, r)
		return
	}

	v := validator.New()

	access_token, err := ser.createToken(user, 10*time.Minute)
	if err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}

	refresh_token, err := ser.createToken(user, 7*24*time.Hour)
	if err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}
	user.Refresh_Token = &refresh_token

	err = ser.models.Users.Patch(&user)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrDuplicateEmail):
			v.AddError("email", "a user with this email already exists.")
			ser.failedValidationResponse(w, r, v.Errors)
		case errors.Is(err, database.ErrDuplicatePhoneNumber):
			v.AddError("phone number", "a user with this phone number already exists.")
			ser.failedValidationResponse(w, r, v.Errors)
		default:
			ser.serverErrorResponse(w, r, err)
		}
		return
	}

	ser.writeJSON(w, http.StatusOK, envelope{"access_token": access_token, "refresh_token": refresh_token}, nil)

}

func (ser Server) refreshToken(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Refresh_Token string `json:"refresh_token"`
	}

	if err := ser.readJSON(w, r, &input); err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}
	user, err := ser.models.Users.GetByRefreshToken(input.Refresh_Token)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrRecordNotFound):
			ser.notFoundResponse(w, r)
		default:
			ser.serverErrorResponse(w, r, err)
		}
		return
	}
	_, err = ser.verifyToken(input.Refresh_Token)
	if err != nil {
		ser.authenticationRequiredResponse(w, r)
		return
	}
	access_token, err := ser.createToken(*user, 10*time.Minute)
	if err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}
	ser.writeJSON(w, http.StatusOK, envelope{"access_token": access_token}, nil)
}

func (ser Server) activate(w http.ResponseWriter, r *http.Request) {

	var input struct {
		PlainText string `json:"plain_text"`
	}

	err := ser.readJSON(w, r, &input)
	if err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}

	hash := sha256.Sum256([]byte(input.PlainText))

	user, err := ser.models.Users.GetToken(hash, database.ScopeActivation, time.Now())
	if err != nil {
		switch {
		case errors.Is(err, database.ErrRecordNotFound):
			ser.forbiddenErrorResponse(w, r)
			return
		default:
			ser.serverErrorResponse(w, r, err)
			return
		}
	}

	user.Is_Verified = true

	err = ser.models.Users.Patch(user)

	if err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}

	if err = ser.models.Tokens.DeleteAllForUser(database.ScopeActivation, user.Id); err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}

	err = ser.writeJSON(w, http.StatusOK, envelope{"user_activation": user.Is_Verified}, nil)
	if err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}
}

func (ser Server) forgotPassword(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email string `json:"email"`
	}

	if err := ser.readJSON(w, r, &input); err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}

	user, err := ser.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrRecordNotFound):
			ser.notFoundResponse(w, r)
			return
		default:
			ser.serverErrorResponse(w, r, err)
			return
		}
	}

	token, err := ser.models.Tokens.New(user.Id, 30*time.Minute, database.ScopePasswordReset)
	if err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}

	go func() {

		data := map[string]interface{}{
			"frontEndURL": "http://autosphere.com",
			"resetToken":  token.Plaintext,
		}

		err := ser.mailer.Send(user.Email, "password_reset.tmpl", data)

		if err != nil {
			ser.logger.PrintError(err, nil)
			return
		}

	}()

	err = ser.writeJSON(w, http.StatusCreated, envelope{"success": true}, nil)
	if err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}

}

func (ser Server) checkTokenExpiry(w http.ResponseWriter, r *http.Request) {

	var input struct {
		PlainText string `json:"plain_text"`
	}

	if err := ser.readJSON(w, r, &input); err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}

	hash := sha256.Sum256([]byte(input.PlainText))
	_, err := ser.models.Users.GetToken(hash, database.ScopePasswordReset, time.Now())

	if err != nil {
		switch {
		case errors.Is(err, database.ErrRecordNotFound):
			ser.writeJSON(w, http.StatusOK, envelope{"expired": "true"}, nil)
			return
		default:
			ser.serverErrorResponse(w, r, err)
			return
		}
	}

	ser.writeJSON(w, http.StatusOK, envelope{"expired": "false"}, nil)

}

func (ser Server) resetPassword(w http.ResponseWriter, r *http.Request) {
	var input struct {
		PlainText   string `json:"plain_text"`
		NewPassword string `json:"new_password"`
	}

	if err := ser.readJSON(w, r, &input); err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}

	hash := sha256.Sum256([]byte(input.PlainText))
	user, err := ser.models.Users.GetToken(hash, database.ScopePasswordReset, time.Now())

	if err != nil {
		switch {
		case errors.Is(err, database.ErrRecordNotFound):
			ser.writeJSON(w, http.StatusBadRequest, envelope{"success": false}, nil)
			return
		default:
			ser.serverErrorResponse(w, r, err)
			return
		}
	}

	pass := &user.Password

	err = pass.Set(input.NewPassword)
	if err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}

	err = ser.models.Users.Patch(user)
	if err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}

	if err = ser.models.Tokens.DeleteAllForUser(database.ScopePasswordReset, user.Id); err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}

	ser.writeJSON(w, http.StatusOK, envelope{}, nil)

}
