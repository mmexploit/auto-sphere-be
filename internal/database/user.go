package database

import (
	"database/sql"
	"time"
)

type Role int

const (
	ADMIN Role = iota
	OPERATOR
	SELLER
)

type User struct {
	Id           int       `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	Password     string    `json:"-"`
	Phone_Number string    `json:"phone_number"`
	Role         Role      `json:"role"`
	Created_At   time.Time `json:"-"`
}

type UserModel struct {
	db *sql.DB
}

func (um UserModel) Create(user User) error {
	const query = `INSERT INTO users 
				   (name, email, password, phone_number, role) 
				   VALUES ($1,$2,$3,$4,$5)
				   RETURNING id, name, email, phone_number, role`
	args := []interface{}{user.Name, user.Email, user.Password, user.Phone_Number, user.Role}
	return um.db.QueryRow(query, args...).Scan(&user.Id, &user.Name, &user.Email, &user.Phone_Number, &user.Role)
}

func (ser UserModel) get() error {

	return nil
}
func (ser UserModel) update(id string, user UserModel) error {
	return nil
}

func (ser UserModel) delete(id int) error {
	return nil
}
