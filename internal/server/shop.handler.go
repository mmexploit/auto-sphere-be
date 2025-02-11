package server

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Mahider-T/autoSphere/internal/database"
	"github.com/Mahider-T/autoSphere/validator"
)

func (ser *Server) shopCreate(w http.ResponseWriter, r *http.Request) {

	userId := r.Context().Value(UserIDKey).(float64)

	var input struct {
		Name            string                      `json:"name"`
		Phone_Number    string                      `json:"phone_number"`
		Email           string                      `json:"email"`
		Location        string                      `json:"location"`
		Coordinate      string                      `json:"coordinate"`
		Category        []string                    `json:"category"`
		Thumbnail       string                      `json:"thumbnail"`
		Photos          []string                    `json:"photos"`
		Approval_Status database.ShopApprovalStatus `json:"approval_status"`
	}

	if err := ser.readJSON(w, r, &input); err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}

	shop := database.Shop{
		Name:            input.Name,
		Phone_Number:    input.Phone_Number,
		Email:           input.Email,
		Location:        input.Location,
		Coordinate:      input.Coordinate,
		Thumbnail:       &input.Thumbnail,
		Photos:          input.Photos,
		Approval_Status: input.Approval_Status,
		Created_By:      int(userId),
	}

	if err := ser.models.Shops.Create(&shop); err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/shops/%d", shop.Id))
	err := ser.writeJSON(w, http.StatusCreated, envelope{"shop": shop}, headers)
	if err != nil {
		ser.serverErrorResponse(w, r, err)
	}
}

func (ser Server) shopGetOne(w http.ResponseWriter, r *http.Request) {
	id, err := ser.readIDParam(r)
	if err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}

	shop, err := ser.models.Shops.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrRecordNotFound):
			ser.notFoundResponse(w, r)
		default:
			ser.serverErrorResponse(w, r, err)
		}
		return
	}

	err = ser.writeJSON(w, http.StatusOK, envelope{"shop": shop}, nil)
	if err != nil {
		ser.serverErrorResponse(w, r, err)
	}
}

func (ser Server) shopDelete(w http.ResponseWriter, r *http.Request) {
	id, err := ser.readIDParam(r)
	if err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}

	if err = ser.models.Shops.Delete(id); err != nil {
		switch {
		case errors.Is(err, database.ErrRecordNotFound):
			ser.notFoundResponse(w, r)
		default:
			ser.serverErrorResponse(w, r, err)
		}
		return
	}

	err = ser.writeJSON(w, http.StatusOK, envelope{"message": "shop successfully deleted"}, nil)
	if err != nil {
		ser.serverErrorResponse(w, r, err)
	}
}

func (ser Server) shopPatch(w http.ResponseWriter, r *http.Request) {
	id, err := ser.readIDParam(r)
	if err != nil {
		ser.notFoundResponse(w, r)
		return
	}

	shop, err := ser.models.Shops.Get(id)
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
		Name         *string   `json:"name"`
		Phone_Number *string   `json:"phone_number"`
		Email        *string   `json:"email"`
		Location     *string   `json:"location"`
		Coordinate   *string   `json:"coordinate"`
		Category     *[]string `json:"category"`
		Thumbnail    *string   `json:"thumbnail"`
		Photos       *[]string `json:"photos"`
	}

	err = ser.readJSON(w, r, &input)
	if err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}

	if input.Name != nil {
		shop.Name = *input.Name
	}
	if input.Phone_Number != nil {
		shop.Phone_Number = *input.Phone_Number
	}
	if input.Email != nil {
		shop.Email = *input.Email
	}
	if input.Location != nil {
		shop.Location = *input.Location
	}
	if input.Photos != nil {
		shop.Photos = *input.Photos
	}
	if input.Coordinate != nil {

		shop.Coordinate = *input.Coordinate
	} else {
		coords := shop.Coordinate
		if coords != "" {
			// Strip "POINT()" and spaces, extracting just the longitude and latitude
			coords = coords[6 : len(coords)-1]
			// Now coords contains the "longitude latitude" format
			shop.Coordinate = coords
		}
	}

	err = ser.models.Shops.Patch(shop)
	if err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}

	err = ser.writeJSON(w, http.StatusOK, envelope{"shop": shop}, nil)
	if err != nil {
		ser.serverErrorResponse(w, r, err)
	}
}
func (ser Server) getShops(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name             string
		Coordinate       string
		Max_Distance     int
		Category_Members []string
		Filters          database.Filters
	}

	v := validator.New()

	// Parse query parameters
	input.Name = ser.parseString(r, "name", "")
	input.Coordinate = ser.parseString(r, "coordinate", "")
	input.Max_Distance = ser.parseInt(r, "max_dist", 10, v)

	input.Filters.Page = ser.parseInt(r, "page", 1, v)
	input.Filters.PageSize = ser.parseInt(r, "page_size", 20, v)
	input.Filters.Sort = ser.parseString(r, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "name", "created_at", "-id", "-name", "-created_at"}

	category_members, err := ser.parseCSV(r, "category_members", []string{})
	if err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}
	input.Category_Members = category_members
	// fmt.Print("Category members are :- ", input.Category_Members)
	// Validate filters
	if database.ValidateFilters(v, input.Filters); !v.Valid() {
		ser.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Fetch shops
	shops, metadata, err := ser.models.Shops.GetAll(input.Name, input.Coordinate, input.Max_Distance, input.Filters, input.Category_Members)
	if err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}

	// Return JSON response
	ser.writeJSON(w, http.StatusOK, envelope{"metadata": metadata, "shops": shops}, nil)
}

func (ser Server) updateAppoval(w http.ResponseWriter, r *http.Request) {

	id, err := ser.readIDParam(r)
	if err != nil {
		ser.notFoundResponse(w, r)
		return
	}

	var input struct {
		Approval_Status database.ShopApprovalStatus `json:"approval_status"`
	}

	if err := ser.readJSON(w, r, &input); err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}

	if err := ser.models.Shops.UpdateAppoval(id, input.Approval_Status); err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}

	ser.writeJSON(w, http.StatusOK, envelope{"id": id, "approval_status": input.Approval_Status}, nil)

}
