package server

import (
	"net/http"

	"github.com/Mahider-T/autoSphere/internal/database"
)

func (ser Server) scCreate(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Shop_id            int `json:"shop_id"`
		Category_member_id int `json:"category_member_id"`
	}

	if err := ser.readJSON(w, r, &input); err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}

	sc := database.ShopCategory{
		Shop_id:            input.Shop_id,
		Category_member_id: input.Category_member_id,
	}
	if err := ser.models.ShopCategory.Create(sc); err != nil {
		ser.serverErrorResponse(w, r, err)
		return
	}

	ser.writeJSON(w, http.StatusCreated, envelope{"shop_category": sc}, nil)
}
