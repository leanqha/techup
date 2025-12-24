package account

import "time"

type Account struct {
	ID           int
	UID          string
	Email        string
	PasswordHash string
	FirstName    string
	MiddleName   string
	LastName     string
	Role         string
	IsVerified   bool
	GroupID      int
	GroupName    string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type RefreshToken struct {
	ID        int       `db:"id"`
	AccountID int       `db:"account_id"`
	Token     string    `db:"token"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
}
