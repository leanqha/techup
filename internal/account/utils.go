package account

import (
	"golang.org/x/crypto/bcrypt"
	"techup/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte(config.GetJWTSecret())

func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GenerateJWT(acc *Account) (string, error) {
	expireTime := time.Now().Add(config.GetJWTAccessTTL())
	claims := jwt.MapClaims{
		"user_id": acc.ID,
		"role":    acc.Role,
		"exp":     expireTime.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.GetJWTSecret()))
}
