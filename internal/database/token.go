package database

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"time"

	"github.com/Mahider-T/autoSphere/validator"
)

const (
	ScopeActivation    = "activation"
	ScopePasswordReset = "password_reset"
)

type Token struct {
	Plaintext string
	Hash      []byte
	UserId    int64
	Expiry    time.Time
	Scope     string
}

func generateToken(userId int64, ttl time.Duration, scope string) (*Token, error) {

	token := &Token{
		UserId: userId,
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}

	randomBytes := make([]byte, 16)

	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)

	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]
	return token, nil
}

func ValidateTokenPlaintext(v *validator.Validator, tokenPlaintext string) {
	v.Check(tokenPlaintext != "", "token", "must be provided")
	v.Check(len(tokenPlaintext) == 26, "token", "must be 26 bytes long")
}

type TokenModel struct {
	db *sql.DB
}

func (m TokenModel) New(userId int64, ttl time.Duration, scope string) (*Token, error) {
	token, err := generateToken(userId, ttl, scope)
	if err != nil {
		return nil, err
	}
	err = m.Insert(token)
	return token, err
}

func (m TokenModel) Insert(token *Token) error {
	query := `
	INSERT INTO tokens (hash, user_id, expiry, scope)
	VALUES ($1, $2, $3, $4)`
	args := []interface{}{token.Hash, token.UserId, token.Expiry, token.Scope}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := m.db.ExecContext(ctx, query, args...)
	return err
}

func (m TokenModel) DeleteAllForUser(scope string, userId int64) error {
	query := `
	DELETE FROM tokens
	WHERE scope = $1 AND user_id = $2`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := m.db.ExecContext(ctx, query, scope, userId)
	return err
}
