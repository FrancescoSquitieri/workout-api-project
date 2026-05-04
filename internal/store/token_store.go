package store

import (
	"apiProject/internal/tokens"
	"database/sql"
	"time"
)

type PostgresTokenStore struct {
	db *sql.DB
}

func NewPostgresTokenStore(db *sql.DB) *PostgresTokenStore {
	return &PostgresTokenStore{
		db: db,
	}
}

type TokenStore interface {
	Insert(token *tokens.Token) error
	CreateNewToken(userId int, ttl time.Duration, scope string) (*tokens.Token, error)
	DeleteAllTokensForUser(userId int, scope string) error
}

func (pdb *PostgresTokenStore) CreateNewToken(userId int, ttl time.Duration, scope string) (*tokens.Token, error) {
	token, err := tokens.GenerateToken(userId, ttl, scope)
	if err != nil {
		return nil, err
	}

	err = pdb.Insert(token)
	return token, err
}

func (pdb *PostgresTokenStore) Insert(token *tokens.Token) error {
	query := `
	INSERT INTO tokens (hash, user_id, expiry, scope)
	VALUES ($1, $2, $3, $4)
`
	_, err := pdb.db.Exec(query, token.Hash, token.UserID, token.Expiry, token.Scope)
	if err != nil {
		return err
	}
	return nil
}

func (pdb *PostgresTokenStore) DeleteAllTokensForUser(userId int, scope string) error {
	query := `
		DELETE INTO tokens
		WHERE user_id = $1 AND scope = $2
	`
	_, err := pdb.db.Exec(query, userId, scope)
	return err
}
