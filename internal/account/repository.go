package account

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"techup/internal/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateAccount(ctx context.Context, acc *Account) error {
	query := `
		INSERT INTO accounts (email, password_hash, first_name, last_name, role)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	err := r.db.QueryRow(ctx, query, acc.Email, acc.PasswordHash, acc.FirstName, acc.LastName, acc.Role).Scan(&acc.ID)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return fmt.Errorf("email already exists")
		}
		logger.LogSQLError(err, query, acc.Email, acc.PasswordHash)
		return err
	}
	return nil
}

func (r *Repository) GetByEmail(ctx context.Context, email string) (*Account, error) {
	acc := &Account{}
	query := `SELECT id, email, password_hash, first_name, last_name, role FROM accounts WHERE email=$1`
	err := r.db.QueryRow(ctx, query, email).Scan(&acc.ID, &acc.Email, &acc.PasswordHash, &acc.FirstName, &acc.LastName, &acc.Role)
	if err != nil {
		logger.LogSQLError(err, query, email)
		return nil, fmt.Errorf("account not found")
	}
	return acc, nil
}

func (r *Repository) GetByID(ctx context.Context, id int) (*Account, error) {
	acc := &Account{}
	query := `SELECT id, uid, email, password_hash, first_name, last_name, role 
	          FROM accounts WHERE id=$1`
	err := r.db.QueryRow(ctx, query, id).Scan(
		&acc.ID, &acc.UID, &acc.Email, &acc.PasswordHash,
		&acc.FirstName, &acc.LastName, &acc.Role,
	)
	if err != nil {
		logger.LogSQLError(err, query, id)
		return nil, err
	}
	return acc, nil
}

func (r *Repository) UpdateAccount(ctx context.Context, acc *Account) error {
	query := `
		UPDATE accounts
		SET email = $1,
		    first_name = $2,
		    last_name = $3,
		    password_hash = $4,
		    role = $5,
		    updated_at = NOW()
		WHERE id = $6
	`
	commandTag, err := r.db.Exec(ctx, query,
		acc.Email,
		acc.FirstName,
		acc.LastName,
		acc.PasswordHash,
		acc.Role,
		acc.ID,
	)
	if err != nil {
		logger.LogSQLError(err, query, acc.Email, acc.ID)
		return err
	}
	if commandTag.RowsAffected() != 1 {
		return fmt.Errorf("account not found or not updated")
	}
	return nil
}

func (r *Repository) SaveRefreshToken(ctx context.Context, token *RefreshToken) error {
	query := `
        INSERT INTO refresh_tokens (account_id, token, expires_at)
        VALUES ($1, $2, $3)
    `
	_, err := r.db.Exec(ctx, query, token.AccountID, token.Token, token.ExpiresAt)
	if err != nil {
		logger.LogSQLError(err, query, token.AccountID, token.Token)
		return err
	}
	return nil
}

func (r *Repository) GetRefreshToken(ctx context.Context, token string) (*RefreshToken, error) {
	query := `
        SELECT id, account_id, token, expires_at, created_at
        FROM refresh_tokens
        WHERE token = $1
    `
	row := r.db.QueryRow(ctx, query, token)
	var t RefreshToken
	err := row.Scan(&t.ID, &t.AccountID, &t.Token, &t.ExpiresAt, &t.CreatedAt)
	if err != nil {
		logger.LogSQLError(err, query, token)
		return nil, err
	}
	return &t, nil
}

func (r *Repository) DeleteRefreshToken(ctx context.Context, oldToken string) error {
	query := `DELETE FROM refresh_tokens WHERE token = $1`
	_, err := r.db.Exec(ctx, query, oldToken)
	if err != nil {
		logger.LogSQLError(err, query, oldToken)
		return err
	}
	return nil
}

func (r *Repository) DeleteRefreshTokens(ctx context.Context, userID int) error {
	_, err := r.db.Exec(ctx, `
		DELETE FROM refresh_tokens
		WHERE account_id = $1
	`, userID)
	return err
}

func (r *Repository) DeleteAccount(ctx context.Context, id int) error {
	if err := r.DeleteRefreshTokens(ctx, id); err != nil {
		return err
	}

	query := `DELETE FROM accounts WHERE id = $1`
	ct, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if ct.RowsAffected() != 1 {
		return errors.New("account not found")
	}
	return nil
}
