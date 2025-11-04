package account

import (
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"techup/config"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetByID(ctx context.Context, id int) (*Account, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Register(ctx context.Context, email, password, firstName, lastName string) (*Account, error) {
	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}
	acc := &Account{
		Email:        email,
		PasswordHash: hash,
		FirstName:    firstName,
		LastName:     lastName,
		Role:         "student",
	}
	err = s.repo.CreateAccount(ctx, acc)
	if err != nil {
		return nil, err
	}
	return acc, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (string, string, error) {
	acc, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return "", "", errors.New("invalid credentials")
	}

	if !CheckPasswordHash(password, acc.PasswordHash) {
		return "", "", errors.New("invalid credentials")
	}

	accessToken, err := GenerateJWT(acc)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := s.CreateRefreshToken(ctx, acc.ID)
	if err != nil {
		return "", "", err
	}

	return accessToken, *refreshToken, nil
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
	return s.repo.Update(ctx, acc)
}

func (s *Service) UpdateProfile(ctx context.Context, userID int, req *UpdateProfileRequest) (*Account, error) {
	acc, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	// проверка уникальности email и обновление
	acc.Email = req.Email
	acc.FirstName = req.FirstName
	acc.LastName = req.LastName
	if err := s.repo.Update(ctx, acc); err != nil {
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
	return s.repo.Update(ctx, acc)
}

func (s *Service) CreateRefreshToken(ctx context.Context, userID int) (*string, error) {
	token := uuid.NewString()
	expiresAt := time.Now().Add(config.GetJWTRefreshTTL())

	rt := &RefreshToken{
		AccountID: userID,
		Token:     token,
		ExpiresAt: expiresAt,
	}

	if err := s.repo.SaveRefreshToken(ctx, rt); err != nil {
		return nil, err
	}

	return &token, nil
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

	// issue new access token
	newAccess, err := GenerateJWT(acc)
	if err != nil {
		return "", "", err
	}

	newRefresh, err := s.CreateRefreshToken(ctx, acc.ID)
	if err != nil {
		return "", "", err
	}

	// delete old refresh
	_ = s.repo.DeleteRefreshToken(ctx, oldToken)

	return newAccess, *newRefresh, nil
}
