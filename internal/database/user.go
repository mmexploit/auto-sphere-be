package database

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Role string

const (
	ADMIN    Role = "ADMIN"
	OPERATOR Role = "OPERATOR"
	SALES    Role = "SALES"
)

type User struct {
	Id           int64     `json:"id"`
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

func (um UserModel) Create(user *User) error {
	const query = `INSERT INTO users 
				   (name, email, password, phone_number, role) 
				   VALUES ($1,$2,$3,$4,$5)
				   RETURNING id, name, email, phone_number, role`
	args := []interface{}{user.Name, user.Email, user.Password, user.Phone_Number, user.Role}
	return um.db.QueryRow(query, args...).Scan(&user.Id, &user.Name, &user.Email, &user.Phone_Number, &user.Role)
}

// func (um UserModel) GetAll(limit, skip int) error {

// const query = `SELECT
// 			   		name, email, phone_number, role, created_at
// 			   from users ORDER BY
// 			   		created_ad DESC
// 				LIMIT=$1
// 				OFFSET=$2 `

// 	return nil

// }

func (ser UserModel) Get(id int64) (*User, error) {

	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `SELECT id, name, email, phone_number, role FROM users WHERE id=$1`

	var user User

	err := ser.db.QueryRow(query, id).Scan(
		&user.Id, &user.Name, &user.Email, &user.Phone_Number, &user.Role,
	)
	fmt.Print("Error 1", err)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}

	}
	fmt.Print("user from repo", user)
	return &user, nil
}
func (ser UserModel) Patch(user *User) error {

	fmt.Println("User is", user)

	query := `UPDATE users SET name=$1, email=$2, phone_number=$3, role=$4 WHERE id=$5 RETURNING id, name, email, phone_number, role`

	args := []interface{}{
		user.Name,
		user.Email,
		user.Phone_Number,
		user.Role,
		user.Id,
	}

	return ser.db.QueryRow(query, args...).Scan(&user.Id, &user.Name, &user.Email, &user.Phone_Number, &user.Role)
}

func (ser UserModel) Delete(id int64) error {

	if id < 1 {
		return ErrRecordNotFound
	}
	query := `DELETE FROM users WHERE id=$1`

	result, err := ser.db.Exec(query, id)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}
