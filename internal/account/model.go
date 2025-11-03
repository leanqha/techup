package account

import "time"

type Account struct {
	ID           int
	Email        string
	PasswordHash string
	FirstName    string
	LastName     string
	Role         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
