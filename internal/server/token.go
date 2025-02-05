package server

import (
	"fmt"
	"time"

	"github.com/Mahider-T/autoSphere/internal/database"
	"github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte("secret-key")

func (ser Server) createToken(user database.User, ttl time.Duration) (string, error) {
	// expirationTime := time.Now().Add(time.Hour * time.Duration(hourMultiplier))
	// fmt.Println("Expires in : - - - ", time.Hour*hourMultiplier)
	expirationTime := time.Now().Add(ttl)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"sub":  user.Id,
			"role": user.Role,
			"exp":  expirationTime.Unix(),
		})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (ser Server) verifyToken(tokenString string) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		return err
	}

	if !token.Valid {
		return fmt.Errorf("invalid token")
	}

	return nil
}
