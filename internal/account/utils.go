package account

import (
	"errors"
	"os"
	"techup/config"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/golang-jwt/jwt/v5"
)

type TokenClaims struct {
	UserID int    `json:"user_id"`
	Role   string `json:"role"`
	UID    string `json:"uid"`
	jwt.RegisteredClaims
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GenerateTokens(acc *Account) (string, string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", "", errors.New("secret env variable not set")
	}

	accessTTL := time.Duration(config.GetAccessTokenTTLSeconds()) * time.Second
	refreshTTL := time.Duration(config.GetRefreshTokenTTLSeconds()) * time.Second

	// --- Access token ---
	accessClaims := TokenClaims{
		UserID: acc.ID,
		Role:   acc.Role,
		UID:    acc.UID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).
		SignedString([]byte(secret))
	if err != nil {
		return "", "", err
	}

	// --- Refresh Token ---
	refreshClaims := TokenClaims{
		UserID: acc.ID,
		Role:   acc.Role,
		UID:    acc.UID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(refreshTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).
		SignedString([]byte(secret))
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// ParseToken parses and validates JWT
func ParseToken(tokenStr string) (jwt.MapClaims, error) {
	secret := config.GetJWTSecret()

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		return nil, err
	}

	return token.Claims.(jwt.MapClaims), nil
}
