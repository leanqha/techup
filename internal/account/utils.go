package account

import (
	"errors"
	"fmt"
	"techup/config"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type TokenClaims struct {
	UserID int    `json:"user_id"`
	Role   string `json:"role"`
	UID    string `json:"uid"`
	Type   string `json:"type"`
	jwt.RegisteredClaims
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GenerateTokens(acc *Account) (string, string, error) {
	secret := config.GetJWTSecret()
	if secret == "" {
		return "", "", errors.New("JWT secret not set")
	}

	accessTTL := time.Duration(config.GetAccessTokenTTLSeconds()) * time.Second
	refreshTTL := time.Duration(config.GetRefreshTokenTTLSeconds()) * time.Second

	now := time.Now()

	accessClaims := TokenClaims{
		UserID: acc.ID,
		Role:   acc.Role,
		UID:    acc.UID,
		Type:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(accessTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).
		SignedString([]byte(secret))
	if err != nil {
		return "", "", err
	}

	refreshClaims := TokenClaims{
		UserID: acc.ID,
		Role:   acc.Role,
		UID:    acc.UID,
		Type:   "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(refreshTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
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

// GetUserIDFromContext extracts the user ID from gin.Context.
func GetUserIDFromContext(c *gin.Context) (int, error) {
	uid, exists := c.Get("user_id")
	if !exists {
		return 0, fmt.Errorf("user_id not found in context")
	}
	userID, ok := uid.(int)
	if !ok {
		return 0, fmt.Errorf("user_id in context is not of type int")
	}
	return userID, nil
}
