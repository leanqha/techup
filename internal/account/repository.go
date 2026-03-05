package account

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"techup/internal/logger"

	"github.com/jackc/pgx/v5"
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
	var middleName sql.NullString

	query := `SELECT id, email, password_hash, first_name, middle_name, last_name, role 
	          FROM accounts WHERE email=$1`
	err := r.db.QueryRow(ctx, query, email).Scan(
		&acc.ID, &acc.Email, &acc.PasswordHash,
		&acc.FirstName, &middleName, &acc.LastName, &acc.Role,
	)
	if err != nil {
		logger.LogSQLError(err, query, email)
		return nil, fmt.Errorf("account not found")
	}

	if middleName.Valid {
		acc.MiddleName = middleName.String
	}

	return acc, nil
}

func (r *Repository) GetByID(ctx context.Context, id int) (*Account, error) {
	acc := &Account{}
	var middleName sql.NullString
	var groupName sql.NullString

	query := `
		SELECT a.id, a.uid, a.email, a.password_hash, a.first_name, a.middle_name, a.last_name, a.role, a.is_verified, g.name, a.group_id
		FROM accounts a
		LEFT JOIN groups g ON a.group_id = g.id
		WHERE a.id = $1
	`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&acc.ID, &acc.UID, &acc.Email, &acc.PasswordHash,
		&acc.FirstName, &middleName, &acc.LastName, &acc.Role,
		&acc.IsVerified, &groupName, &acc.GroupID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("account not found")
		}
		logger.LogSQLError(err, query, id)
		return nil, err
	}

	if middleName.Valid {
		acc.MiddleName = middleName.String
	}
	if groupName.Valid {
		acc.GroupName = groupName.String
	}

	return acc, nil
}

func (r *Repository) UpdateAccount(ctx context.Context, acc *Account) error {
	query := `
		UPDATE accounts
		SET uid = $1,
		    email = $2,
		    first_name = $3,
		    middle_name = $4,
		    last_name = $5,
		    password_hash = $6,
		    role = $7,
		    group_id = $8,
		    is_verified = $9,
		    updated_at = NOW()
		WHERE id = $10
	`

	commandTag, err := r.db.Exec(ctx, query,
		acc.UID,
		acc.Email,
		acc.FirstName,
		acc.MiddleName,
		acc.LastName,
		acc.PasswordHash,
		acc.Role,
		acc.GroupID,
		acc.IsVerified,
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

func (r *Repository) CreatePasswordResetToken(ctx context.Context, token *PasswordResetToken) error {
	query := `
		INSERT INTO password_reset_tokens (account_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	err := r.db.QueryRow(ctx, query, token.AccountID, token.TokenHash, token.ExpiresAt).
		Scan(&token.ID, &token.CreatedAt)
	if err != nil {
		logger.LogSQLError(err, query, token.AccountID)
		return err
	}
	return nil
}

func (r *Repository) GetPasswordResetTokenByHash(ctx context.Context, tokenHash string) (*PasswordResetToken, error) {
	query := `
		SELECT id, account_id, token_hash, expires_at, used_at, created_at
		FROM password_reset_tokens
		WHERE token_hash = $1
	`
	row := r.db.QueryRow(ctx, query, tokenHash)
	var t PasswordResetToken
	var usedAt sql.NullTime
	if err := row.Scan(&t.ID, &t.AccountID, &t.TokenHash, &t.ExpiresAt, &usedAt, &t.CreatedAt); err != nil {
		logger.LogSQLError(err, query, tokenHash)
		return nil, err
	}
	if usedAt.Valid {
		t.UsedAt = &usedAt.Time
	}
	return &t, nil
}

func (r *Repository) MarkPasswordResetTokenUsed(ctx context.Context, tokenID int) error {
	query := `
		UPDATE password_reset_tokens
		SET used_at = NOW()
		WHERE id = $1 AND used_at IS NULL
	`
	ct, err := r.db.Exec(ctx, query, tokenID)
	if err != nil {
		logger.LogSQLError(err, query, tokenID)
		return err
	}
	if ct.RowsAffected() != 1 {
		return fmt.Errorf("password reset token not found")
	}
	return nil
}

func (r *Repository) DeletePasswordResetTokens(ctx context.Context, accountID int) error {
	query := `DELETE FROM password_reset_tokens WHERE account_id = $1`
	_, err := r.db.Exec(ctx, query, accountID)
	if err != nil {
		logger.LogSQLError(err, query, accountID)
		return err
	}
	return nil
}

func (r *Repository) ListAccounts(ctx context.Context, f AdminAccountsFilter) ([]Account, error) {
	query := `
		SELECT a.id, a.uid, a.email, a.first_name, a.middle_name, a.last_name, a.role, a.is_verified, a.group_id, g.name
		FROM accounts a
		LEFT JOIN groups g ON a.group_id = g.id
	`

	var conditions []string
	var args []any
	i := 1

	if f.Role != nil {
		conditions = append(conditions, fmt.Sprintf("a.role = $%d", i))
		args = append(args, *f.Role)
		i++
	}
	if f.GroupID != nil {
		conditions = append(conditions, fmt.Sprintf("a.group_id = $%d", i))
		args = append(args, *f.GroupID)
		i++
	}
	if f.IsVerified != nil {
		conditions = append(conditions, fmt.Sprintf("a.is_verified = $%d", i))
		args = append(args, *f.IsVerified)
		i++
	}
	if f.Email != nil {
		conditions = append(conditions, fmt.Sprintf("a.email ILIKE $%d", i))
		args = append(args, "%"+*f.Email+"%")
		i++
	}
	if f.UID != nil {
		conditions = append(conditions, fmt.Sprintf("a.uid ILIKE $%d", i))
		args = append(args, "%"+*f.UID+"%")
		i++
	}
	if f.Name != nil {
		conditions = append(conditions, fmt.Sprintf("(a.first_name ILIKE $%d OR a.middle_name ILIKE $%d OR a.last_name ILIKE $%d)", i, i, i))
		args = append(args, "%"+*f.Name+"%")
		i++
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY a.id"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []Account
	for rows.Next() {
		var acc Account
		var middleName sql.NullString
		var groupName sql.NullString

		if err := rows.Scan(
			&acc.ID,
			&acc.UID,
			&acc.Email,
			&acc.FirstName,
			&middleName,
			&acc.LastName,
			&acc.Role,
			&acc.IsVerified,
			&acc.GroupID,
			&groupName,
		); err != nil {
			return nil, err
		}

		if middleName.Valid {
			acc.MiddleName = middleName.String
		}
		if groupName.Valid {
			acc.GroupName = groupName.String
		}

		accounts = append(accounts, acc)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return accounts, nil
}
