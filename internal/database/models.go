package database

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
)

type Models struct {
	Users UserModel
	Shops ShopModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Users: UserModel{db: db},
		Shops: ShopModel{db: db},
	}
}
