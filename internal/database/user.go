package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Mahider-T/autoSphere/validator"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail       = errors.New("duplicate email")
	ErrDuplicatePhoneNumber = errors.New("duplicate phone number")
)

type Role string

const (
	ADMIN    Role = "ADMIN"
	OPERATOR Role = "OPERATOR"
	SALES    Role = "SALES"
)

type User struct {
	Id            int64     `json:"id"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	Is_Verified   bool      `json:"is_verified"`
	Password      password  `json:"-"`
	Phone_Number  string    `json:"phone_number"`
	Role          Role      `json:"role"`
	Created_At    time.Time `json:"-"`
	Refresh_Token *string   `json:"refresh_token"`
}

func ValidateUser(v *validator.Validator, u *User) {
	var phoneRegex = regexp.MustCompile(`^(09|07)\d{8}$`)
	var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	v.Check(u.Name != "", "name", "name must not be empty")
	v.Check(u.Email != "", "email", "email must not be empty")
	v.Check(validator.Matches(u.Email, emailRegex), "email", "not a valid email")
	v.Check(u.Phone_Number != "", "phone_number", "phone number must not be empty")
	v.Check(validator.Matches(u.Phone_Number, phoneRegex), "phone_number", "phone number must start with 07 or 09 and must be 10 digits long")
	v.Check(u.Role != "", "role", "role must not be empty")
}

type UserModel struct {
	db *sql.DB
}
type password struct {
	plaintext *string
	hash      []byte
}

func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}
	p.plaintext = &plaintextPassword
	p.hash = hash

	return nil
}

func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

func (um UserModel) Create(user *User) error {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()
	const query = `INSERT INTO users 
				   (name, email, password, phone_number, role) 
				   VALUES ($1,$2,$3,$4,$5)
				   RETURNING id, name, email, phone_number, role`
	args := []interface{}{user.Name, user.Email, user.Password.hash, user.Phone_Number, user.Role}
	err := um.db.QueryRowContext(ctx, query, args...).Scan(&user.Id, &user.Name, &user.Email, &user.Phone_Number, &user.Role)

	if err != nil {
		switch {
		case strings.Contains(err.Error(), "users_email_key"):
			return ErrDuplicateEmail
		case strings.Contains(err.Error(), "users_phone_number_key"):
			return ErrDuplicatePhoneNumber
		default:
			return err
		}
	}
	return nil
}

func (um UserModel) GetByEmail(email string) (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT id, name, email, phone_number, role, password, refresh_token FROM users WHERE email=$1`
	var user User
	err := um.db.QueryRowContext(ctx, query, email).Scan(&user.Id, &user.Name, &user.Email, &user.Phone_Number, &user.Role, &user.Password.hash, &user.Refresh_Token)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return User{}, ErrRecordNotFound
		default:
			return User{}, err
		}
	}
	return user, nil
}

func (um UserModel) Get(id int64) (*User, error) {

	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `SELECT id, name, email, phone_number, role FROM users WHERE id=$1`

	var user User
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	err := um.db.QueryRowContext(ctx, query, id).Scan(
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
	return &user, nil
}
func (um UserModel) Patch(user *User) error {

	// fmt.Print("MR patched user is -   -   ", user)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	query := `UPDATE users SET name=$1, email=$2, phone_number=$3, role=$4, refresh_token=$5, is_verified=$6, password=$7 WHERE id=$8 RETURNING id, name, email, phone_number, role, is_verified`

	args := []interface{}{
		user.Name,
		user.Email,
		user.Phone_Number,
		user.Role,
		user.Refresh_Token,
		user.Is_Verified,
		user.Password.hash,
		user.Id,
	}

	err := um.db.QueryRowContext(ctx, query, args...).Scan(&user.Id, &user.Name, &user.Email, &user.Phone_Number, &user.Role, &user.Is_Verified)

	if err != nil {
		switch {
		case strings.Contains(err.Error(), "users_email_key"):
			return ErrDuplicateEmail
		case strings.Contains(err.Error(), "users_phone_number_key"):
			return ErrDuplicatePhoneNumber
		default:
			return err
		}
	}
	return nil

}

func (um UserModel) Delete(id int64) error {

	if id < 1 {
		return ErrRecordNotFound
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	query := `DELETE FROM users WHERE id=$1`

	result, err := um.db.ExecContext(ctx, query, id)

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

func (um UserModel) GetAll(name string, role string, filters Filters) ([]User, Metadata, error) {

	query := fmt.Sprintf(`SELECT count(*) OVER(), id, name, email, phone_number, role, created_at
			  FROM users
			  WHERE (to_tsvector('simple', name) @@ plainto_tsquery('simple', $1) OR $1 = '')
			  AND ($2 = '' OR role = $2::role)
			  ORDER BY %s %s, id ASC
			  LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	ctx, close := context.WithTimeout(context.Background(), 3*time.Second)
	defer close()
	// rows, err := ser.db.QueryContext(ctx, query)
	rows, err := um.db.QueryContext(ctx, query, name, role, filters.limit(), filters.offset())

	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	users := []User{}
	for rows.Next() {
		var user User

		err := rows.Scan(
			&totalRecords,
			&user.Id,
			&user.Name,
			&user.Email,
			&user.Phone_Number,
			&user.Role,
			&user.Created_At,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}
	metadata := filters.calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return users, metadata, nil
}

func (um UserModel) GetByRefreshToken(refreshToken string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT id, role from users
			  WHERE refresh_token=$1`
	var user User
	err := um.db.QueryRowContext(ctx, query, refreshToken).Scan(&user.Id, &user.Role)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}

	}

	return &user, nil
}
func (um UserModel) GetToken(hashText [32]byte, scope string, expiry time.Time) (*User, error) {

	query := `SELECT id, name, email, password, is_verified, phone_number, role, created_at
			  FROM users 
			  INNER JOIN tokens ON users.Id = tokens.user_id
			  WHERE tokens.scope=$1 AND tokens.hash=$2 AND tokens.expiry >= $3`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{
		scope,
		hashText[:],
		expiry,
	}

	var user User
	err := um.db.QueryRowContext(ctx, query, args...).Scan(&user.Id, &user.Name, &user.Email, &user.Password.hash, &user.Is_Verified, &user.Phone_Number, &user.Role, &user.Created_At)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil

}
