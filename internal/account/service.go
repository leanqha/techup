package account

import "context"

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
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
	if err != nil || !CheckPasswordHash(password, acc.PasswordHash) {
		return "", err
	}
	return GenerateJWT(acc.ID, acc.Role)
}

func (s *Service) GetByID(ctx context.Context, id int) (*Account, error) {
	return s.repo.GetByID(ctx, id)
}
