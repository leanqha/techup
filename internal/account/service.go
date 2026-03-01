package account

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
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
	ListAccounts(ctx context.Context, f AdminAccountsFilter) ([]Account, error)

	SaveRefreshToken(ctx context.Context, token *RefreshToken) error
	GetRefreshToken(ctx context.Context, token string) (*RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, token string) error
	DeleteRefreshTokens(ctx context.Context, userID int) error
	DeleteAccount(ctx context.Context, userID int) error

	CreatePasswordResetToken(ctx context.Context, token *PasswordResetToken) error
	GetPasswordResetTokenByHash(ctx context.Context, tokenHash string) (*PasswordResetToken, error)
	MarkPasswordResetTokenUsed(ctx context.Context, tokenID int) error
	DeletePasswordResetTokens(ctx context.Context, accountID int) error
}

type Service struct {
	repo     RepositoryInterface
	notifier PasswordResetNotifier
}

func NewService(repo RepositoryInterface, notifier ...PasswordResetNotifier) *Service {
	var n PasswordResetNotifier
	if len(notifier) > 0 {
		n = notifier[0]
	}
	return &Service{repo: repo, notifier: n}
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
	acc, err := s.repo.GetByEmail(ctx, strings.TrimSpace(email))
	if err != nil {
		return "", "", errors.New("invalid credentials")
	}

	if bcrypt.CompareHashAndPassword([]byte(acc.PasswordHash), []byte(password)) != nil {
		return "", "", errors.New("invalid credentials")
	}

	accessToken, refreshToken, err := GenerateTokens(acc)
	if err != nil {
		return "", "", err
	}

	err = s.repo.SaveRefreshToken(ctx, &RefreshToken{
		AccountID: acc.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(
			time.Duration(config.GetRefreshTokenTTLSeconds()) * time.Second,
		),
	})
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

func (s *Service) UpdateAccountAdmin(ctx context.Context, userID int, req *AdminUpdateAccountRequest) (*Account, error) {
	acc, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if req.Email != nil {
		email := strings.TrimSpace(*req.Email)
		if email == "" {
			return nil, errors.New("email is required")
		}
		acc.Email = email
	}
	if req.FirstName != nil {
		acc.FirstName = *req.FirstName
	}
	if req.MiddleName != nil {
		acc.MiddleName = *req.MiddleName
	}
	if req.LastName != nil {
		acc.LastName = *req.LastName
	}
	if req.Role != nil {
		role := strings.TrimSpace(*req.Role)
		if role == "" {
			return nil, errors.New("role is required")
		}
		acc.Role = role
	}
	if req.GroupID != nil {
		acc.GroupID = *req.GroupID
	}
	if req.IsVerified != nil {
		acc.IsVerified = *req.IsVerified
	}

	if err := s.repo.UpdateAccount(ctx, acc); err != nil {
		return nil, err
	}

	return acc, nil
}

func (s *Service) ListAccounts(ctx context.Context, f AdminAccountsFilter) ([]Account, error) {
	return s.repo.ListAccounts(ctx, f)
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

func (s *Service) RefreshTokens(ctx context.Context, refreshToken string) (string, string, error) {
	claims, err := ParseToken(refreshToken)
	if err != nil {
		return "", "", errors.New("invalid refresh token")
	}

	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != "refresh" {
		return "", "", errors.New("invalid token type")
	}

	rt, err := s.repo.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		return "", "", errors.New("invalid refresh token")
	}

	if time.Now().After(rt.ExpiresAt) {
		_ = s.repo.DeleteRefreshToken(ctx, refreshToken)
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

	_ = s.repo.DeleteRefreshToken(ctx, refreshToken)

	err = s.repo.SaveRefreshToken(ctx, &RefreshToken{
		AccountID: acc.ID,
		Token:     newRefresh,
		ExpiresAt: time.Now().Add(
			time.Duration(config.GetRefreshTokenTTLSeconds()) * time.Second,
		),
	})
	if err != nil {
		return "", "", err
	}

	return newAccess, newRefresh, nil
}

// Logout clears all sessions
func (s *Service) Logout(ctx context.Context, userID int, refreshToken string) error {
	if refreshToken != "" {
		if err := s.repo.DeleteRefreshToken(ctx, refreshToken); err != nil {
			return err
		}
	}
	if userID > 0 {
		if err := s.repo.DeleteRefreshTokens(ctx, userID); err != nil {
			return err
		}
	}
	if userID <= 0 && refreshToken == "" {
		return errors.New("logout requires user ID or refresh token")
	}
	return nil
}

func (s *Service) DeleteAccount(ctx context.Context, userID int) error {
	return s.repo.DeleteAccount(ctx, userID)
}

func (s *Service) RequestPasswordReset(ctx context.Context, email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return errors.New("email is required")
	}

	acc, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			return nil
		}
		return err
	}

	resetToken, err := generateResetToken()
	if err != nil {
		return err
	}

	tokenHash := hashResetToken(resetToken)
	expiresAt := time.Now().Add(time.Duration(config.GetPasswordResetTokenTTLSeconds()) * time.Second)

	if err := s.repo.DeletePasswordResetTokens(ctx, acc.ID); err != nil {
		return err
	}

	if err := s.repo.CreatePasswordResetToken(ctx, &PasswordResetToken{
		AccountID: acc.ID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}); err != nil {
		return err
	}

	if s.notifier != nil {
		if err := s.notifier.SendPasswordReset(ctx, acc.Email, resetToken); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) ResetPassword(ctx context.Context, token, newPassword string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return errors.New("token is required")
	}
	if len(newPassword) < 8 {
		return errors.New("new password must be at least 8 characters")
	}

	resetToken, err := s.repo.GetPasswordResetTokenByHash(ctx, hashResetToken(token))
	if err != nil {
		return errors.New("invalid or expired token")
	}
	if resetToken.UsedAt != nil || time.Now().After(resetToken.ExpiresAt) {
		return errors.New("invalid or expired token")
	}

	acc, err := s.repo.GetByID(ctx, resetToken.AccountID)
	if err != nil {
		return err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	acc.PasswordHash = string(hashed)
	if err := s.repo.UpdateAccount(ctx, acc); err != nil {
		return err
	}

	if err := s.repo.MarkPasswordResetTokenUsed(ctx, resetToken.ID); err != nil {
		return err
	}

	_ = s.repo.DeleteRefreshTokens(ctx, acc.ID)
	_ = s.repo.DeletePasswordResetTokens(ctx, acc.ID)
	return nil
}

func generateResetToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func hashResetToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
