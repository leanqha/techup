package account

import (
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
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
	hash, _ := HashPassword(password)
	acc := &Account{
		Email:        email,
		PasswordHash: hash,
		FirstName:    firstName,
		LastName:     lastName,
		Role:         "student",
	}
	err := s.repo.CreateAccount(ctx, acc)
	if err != nil {
		return nil, err
	}
	return acc, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (string, error) {
	acc, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	if !CheckPasswordHash(password, acc.PasswordHash) {
		return "", errors.New("invalid credentials")
	}
	return GenerateJWT(acc.ID, acc.Role)
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
