package account

import (
	"context"
	"errors"
	"strings"
	"techup/config"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type RepositoryInterface interface {
	CreateAccount(ctx context.Context, acc *Account) error
	GetByEmail(ctx context.Context, email string) (*Account, error)
	GetByID(ctx context.Context, id int) (*Account, error)
	UpdateAccount(ctx context.Context, acc *Account) error

	SaveRefreshToken(ctx context.Context, token *RefreshToken) error
	GetRefreshToken(ctx context.Context, token string) (*RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, token string) error
	DeleteRefreshTokens(ctx context.Context, userID int) error
	DeleteAccount(ctx context.Context, userID int) error
}

type Service struct {
	repo RepositoryInterface
}

func NewService(repo RepositoryInterface) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetByID(ctx context.Context, id int) (*Account, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*Account, string, string, error) {
	req.Email = strings.TrimSpace(req.Email)
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)
	if req.Email == "" || req.FirstName == "" || req.LastName == "" || len(req.Password) < 6 {
		return nil, "", "", errors.New("invalid input data")
	}

	exists, _ := s.repo.GetByEmail(ctx, req.Email)
	if exists != nil {
		return nil, "", "", errors.New("email already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", "", err
	}

	acc := &Account{
		Email:        req.Email,
		PasswordHash: string(hash),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Role:         "student",
	}

	if err := s.repo.CreateAccount(ctx, acc); err != nil {
		return nil, "", "", err
	}

	accessToken, refreshToken, err := GenerateTokens(acc)
	if err != nil {
		return nil, "", "", err
	}

	if err := s.repo.SaveRefreshToken(ctx, &RefreshToken{
		AccountID: acc.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(time.Duration(config.GetRefreshTokenTTLSeconds()) * time.Second),
	}); err != nil {
		return nil, "", "", err
	}
	return acc, accessToken, refreshToken, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (string, string, error) {
	acc, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return "", "", errors.New("invalid credentials")
	}

	if !CheckPasswordHash(password, acc.PasswordHash) {
		return "", "", errors.New("invalid credentials")
	}

	accessToken, refreshToken, err := GenerateTokens(acc)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *Service) ChangePassword(ctx context.Context, userID int, req *ChangePasswordRequest) error {
	acc, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if bcrypt.CompareHashAndPassword([]byte(acc.PasswordHash), []byte(req.OldPassword)) != nil {
		return errors.New("old password is incorrect")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	acc.PasswordHash = string(hashed)
	return s.repo.UpdateAccount(ctx, acc)
}

func (s *Service) UpdateProfile(ctx context.Context, userID int, req *UpdateProfileRequest) (*Account, error) {
	acc, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	acc.Email = req.Email
	acc.FirstName = req.FirstName
	acc.LastName = req.LastName
	if err := s.repo.UpdateAccount(ctx, acc); err != nil {
		return nil, err
	}
	return acc, nil
}

func (s *Service) SetRole(ctx context.Context, userID int, req *SetRoleRequest) error {
	admin, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if admin.Role != "admin" {
		return errors.New("forbidden")
	}
	acc, err := s.repo.GetByID(ctx, req.UserID)
	if err != nil {
		return err
	}
	acc.Role = req.Role
	return s.repo.UpdateAccount(ctx, acc)
}

func (s *Service) RefreshTokens(ctx context.Context, oldToken string) (string, string, error) {
	rt, err := s.repo.GetRefreshToken(ctx, oldToken)
	if err != nil {
		return "", "", errors.New("invalid refresh token")
	}

	if time.Now().After(rt.ExpiresAt) {
		_ = s.repo.DeleteRefreshToken(ctx, oldToken)
		return "", "", errors.New("refresh token expired")
	}

	acc, err := s.repo.GetByID(ctx, rt.AccountID)
	if err != nil {
		return "", "", err
	}

	newAccess, newRefresh, err := GenerateTokens(acc)
	if err != nil {
		return "", "", err
	}

	_ = s.repo.DeleteRefreshToken(ctx, oldToken)
	err = s.repo.SaveRefreshToken(ctx, &RefreshToken{
		AccountID: acc.ID,
		Token:     newRefresh,
		ExpiresAt: time.Now().Add(time.Duration(config.GetRefreshTokenTTLSeconds()) * time.Second),
	})
	if err != nil {
		return "", "", err
	}

	return newAccess, newRefresh, nil
}

// Logout clears all sessions
func (s *Service) Logout(ctx context.Context, userID int) error {
	if err := s.repo.DeleteRefreshTokens(ctx, userID); err != nil {
		return err
	}
	return nil
}

func (s *Service) DeleteAccount(ctx context.Context, userID int) error {
	return s.repo.DeleteAccount(ctx, userID)
}
