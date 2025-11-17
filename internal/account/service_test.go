package account

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

// --- Mock Repository ---
type MockRepo struct {
	accounts      map[string]*Account
	refreshTokens map[string]*RefreshToken
	saveError     error
	getError      error
	deleteError   error
}

func (m *MockRepo) CreateAccount(_ context.Context, acc *Account) error {
	if m.saveError != nil {
		return m.saveError
	}
	if _, exists := m.accounts[acc.Email]; exists {
		return errors.New("email already exists")
	}
	acc.ID = len(m.accounts) + 1
	m.accounts[acc.Email] = acc
	return nil
}

func (m *MockRepo) GetByEmail(_ context.Context, email string) (*Account, error) {
	if m.getError != nil {
		return nil, m.getError
	}
	acc, ok := m.accounts[email]
	if !ok {
		return nil, errors.New("account not found")
	}
	return acc, nil
}

func (m *MockRepo) GetByID(_ context.Context, id int) (*Account, error) {
	for _, acc := range m.accounts {
		if acc.ID == id {
			return acc, nil
		}
	}
	return nil, errors.New("account not found")
}

func (m *MockRepo) UpdateAccount(_ context.Context, acc *Account) error {
	// Check for email conflict with other accounts
	for email, a := range m.accounts {
		if email == acc.Email && a.ID != acc.ID {
			return errors.New("email already exists")
		}
	}
	// Update the account in the map
	for email, a := range m.accounts {
		if a.ID == acc.ID {
			delete(m.accounts, email)
			break
		}
	}
	m.accounts[acc.Email] = acc
	return nil
}
func (m *MockRepo) SaveRefreshToken(_ context.Context, token *RefreshToken) error {
	if m.saveError != nil {
		return m.saveError
	}
	m.refreshTokens[token.Token] = token
	return nil
}
func (m *MockRepo) GetRefreshToken(_ context.Context, token string) (*RefreshToken, error) {
	t, ok := m.refreshTokens[token]
	if !ok {
		return nil, errors.New("not found")
	}
	return t, nil
}

func (m *MockRepo) DeleteRefreshToken(_ context.Context, token string) error {
	if m.deleteError != nil {
		return m.deleteError
	}
	delete(m.refreshTokens, token)
	return nil
}

func (m *MockRepo) DeleteRefreshTokens(_ context.Context, userID int) error {
	if m.deleteError != nil {
		return m.deleteError
	}
	for token, rt := range m.refreshTokens {
		if rt.AccountID == userID {
			delete(m.refreshTokens, token)
		}
	}
	return nil
}

func (m *MockRepo) DeleteAccount(ctx context.Context, userID int) error {
	if m.deleteError != nil {
		return m.deleteError
	}
	// Удаляем refresh токены пользователя
	for token, rt := range m.refreshTokens {
		if rt.AccountID == userID {
			delete(m.refreshTokens, token)
		}
	}
	// Удаляем сам аккаунт
	for email, acc := range m.accounts {
		if acc.ID == userID {
			delete(m.accounts, email)
			return nil
		}
	}
	return errors.New("account not found")
}

// --- Tests ---
func TestRegisterAndLogin(t *testing.T) {
	repo := &MockRepo{
		accounts:      make(map[string]*Account),
		refreshTokens: make(map[string]*RefreshToken),
	}

	service := NewService(repo)

	// --- Register ---
	acc, accessToken, refreshToken, err := service.Register(context.Background(), RegisterRequest{
		Email:     "test@example.com",
		Password:  "password",
		FirstName: "John",
		LastName:  "Doe",
	})
	assert.NoError(t, err)
	assert.Equal(t, "test@example.com", acc.Email)
	assert.Equal(t, "John", acc.FirstName)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)

	// --- Duplicate email ---
	_, _, _, err = service.Register(context.Background(), RegisterRequest{
		Email:     "test@example.com",
		Password:  "password",
		FirstName: "John",
		LastName:  "Doe",
	})
	assert.Error(t, err)

	// --- Login ---
	access, refresh, err := service.Login(context.Background(), "test@example.com", "password")
	assert.NoError(t, err)
	assert.NotEmpty(t, access)
	assert.NotEmpty(t, refresh)

	// --- Wrong password ---
	_, _, err = service.Login(context.Background(), "test@example.com", "wrongpassword")
	assert.Error(t, err)
}

func TestGetByID(t *testing.T) {
	repo := &MockRepo{
		accounts: map[string]*Account{
			"user@example.com": {ID: 1, Email: "user@example.com", FirstName: "Alice", LastName: "Smith"},
		},
	}

	service := NewService(repo)

	acc, err := service.GetByID(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, "Alice", acc.FirstName)

	acc, err = service.GetByID(context.Background(), 999)
	assert.Error(t, err)
}

func TestRefreshTokens(t *testing.T) {
	repo := &MockRepo{
		accounts:      make(map[string]*Account),
		refreshTokens: make(map[string]*RefreshToken),
	}

	service := NewService(repo)
	acc := &Account{ID: 1, Email: "user@example.com"}
	repo.accounts[acc.Email] = acc

	access, refresh, err := GenerateTokens(acc)
	assert.NoError(t, err)
	assert.NotEmpty(t, access)
	assert.NotEmpty(t, refresh)

	// Save refresh token
	err = service.repo.SaveRefreshToken(context.Background(), &RefreshToken{
		AccountID: acc.ID,
		Token:     refresh,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	})
	assert.NoError(t, err)
}

func TestLogout(t *testing.T) {
	repo := &MockRepo{
		accounts:      make(map[string]*Account),
		refreshTokens: make(map[string]*RefreshToken),
	}

	service := NewService(repo)

	// Создаем пользователя и refresh токен
	acc := &Account{ID: 1, Email: "user@example.com"}
	repo.accounts[acc.Email] = acc

	refreshToken := "refresh-token-123"
	repo.refreshTokens[refreshToken] = &RefreshToken{
		AccountID: acc.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	// Логаут: удалить все токены пользователя
	err := service.Logout(context.Background(), acc.ID)
	assert.NoError(t, err)

	// Проверяем, что токены удалены
	_, exists := repo.refreshTokens[refreshToken]
	assert.False(t, exists)
}

func TestUpdateProfile(t *testing.T) {
	repo := &MockRepo{
		accounts: make(map[string]*Account),
	}

	service := NewService(repo)

	acc := &Account{ID: 1, Email: "user@example.com", FirstName: "John", LastName: "Doe"}
	repo.accounts[acc.Email] = acc

	// Обновляем профиль
	req := UpdateProfileRequest{
		Email:     "new@example.com",
		FirstName: "Jane",
		LastName:  "Smith",
	}
	updatedAcc, err := service.UpdateProfile(context.Background(), acc.ID, &req)
	assert.NoError(t, err)

	assert.Equal(t, "new@example.com", updatedAcc.Email)
	assert.Equal(t, "Jane", updatedAcc.FirstName)
	assert.Equal(t, "Smith", updatedAcc.LastName)

	// Ошибка: обновляем несуществующего пользователя
	_, err = service.UpdateProfile(context.Background(), 999, &req)
	assert.Error(t, err)
}

func TestLoginRefreshLogoutEdgeCases(t *testing.T) {
	repo := &MockRepo{
		accounts:      make(map[string]*Account),
		refreshTokens: make(map[string]*RefreshToken),
	}

	service := NewService(repo)

	// --- Создаем аккаунт с хэшем пароля ---
	hash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	acc := &Account{
		ID:           1,
		Email:        "user@example.com",
		FirstName:    "Alice",
		LastName:     "Smith",
		Role:         "student",
		PasswordHash: string(hash),
	}
	repo.accounts[acc.Email] = acc

	// --- Логин с правильным паролем ---
	access, refresh, err := service.Login(context.Background(), acc.Email, "password")
	assert.NoError(t, err)
	assert.NotEmpty(t, access)
	assert.NotEmpty(t, refresh)

	// --- Логин с неверным паролем ---
	_, _, err = service.Login(context.Background(), acc.Email, "wrongpassword")
	assert.Error(t, err)

	// --- Логин с несуществующим email ---
	_, _, err = service.Login(context.Background(), "missing@example.com", "password")
	assert.Error(t, err)

	// --- Сохраняем refresh токен ---
	err = repo.SaveRefreshToken(context.Background(), &RefreshToken{
		AccountID: acc.ID,
		Token:     refresh,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	})
	assert.NoError(t, err)

	// --- Попытка обновления токенов с недействительным токеном ---
	_, _, err = service.RefreshTokens(context.Background(), "invalid-token")
	assert.Error(t, err)

	// --- Успешный refresh ---
	newAccess, newRefresh, err := service.RefreshTokens(context.Background(), refresh)
	assert.NoError(t, err)
	assert.NotEmpty(t, newAccess)
	assert.NotEmpty(t, newRefresh)

	// --- Логаут ---
	err = service.Logout(context.Background(), acc.ID)
	assert.NoError(t, err)
	for _, rt := range repo.refreshTokens {
		assert.NotEqual(t, acc.ID, rt.AccountID)
	}

	// --- Попытка logout для несуществующего пользователя ---
	err = service.Logout(context.Background(), 999)
	assert.NoError(t, err) // не должно падать, просто ничего не удаляет
}

func TestRegisterValidation(t *testing.T) {
	repo := &MockRepo{accounts: make(map[string]*Account), refreshTokens: make(map[string]*RefreshToken)}
	service := NewService(repo)

	// Empty email
	_, _, _, err := service.Register(context.Background(), RegisterRequest{
		Email: "", Password: "password", FirstName: "A", LastName: "B",
	})
	assert.Error(t, err)

	// Password too short
	_, _, _, err = service.Register(context.Background(), RegisterRequest{
		Email: "a@b.c", Password: "123", FirstName: "A", LastName: "B",
	})
	assert.Error(t, err)

	// Empty first name
	_, _, _, err = service.Register(context.Background(), RegisterRequest{
		Email: "a@b.c", Password: "password", FirstName: "", LastName: "B",
	})
	assert.Error(t, err)

	// Empty last name
	_, _, _, err = service.Register(context.Background(), RegisterRequest{
		Email: "a@b.c", Password: "password", FirstName: "A", LastName: "",
	})
	assert.Error(t, err)
}

func TestPasswordHashing(t *testing.T) {
	repo := &MockRepo{accounts: make(map[string]*Account), refreshTokens: make(map[string]*RefreshToken)}
	service := NewService(repo)

	acc, _, _, err := service.Register(context.Background(), RegisterRequest{
		Email: "hash@test.com", Password: "password", FirstName: "A", LastName: "B",
	})
	assert.NoError(t, err)
	assert.NotEqual(t, "password", acc.PasswordHash)
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(acc.PasswordHash), []byte("password")))
}

func TestRefreshTokenExpired(t *testing.T) {
	repo := &MockRepo{accounts: make(map[string]*Account), refreshTokens: make(map[string]*RefreshToken)}
	service := NewService(repo)
	acc := &Account{ID: 1, Email: "user@test.com"}
	repo.accounts[acc.Email] = acc

	refresh := "expired-token"
	repo.refreshTokens[refresh] = &RefreshToken{
		AccountID: acc.ID,
		Token:     refresh,
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}

	_, _, err := service.RefreshTokens(context.Background(), refresh)
	assert.Error(t, err)
}

func TestRefreshTokenWrongUser(t *testing.T) {
	repo := &MockRepo{accounts: make(map[string]*Account), refreshTokens: make(map[string]*RefreshToken)}
	service := NewService(repo)

	acc := &Account{ID: 1, Email: "real@test.com"}
	repo.accounts[acc.Email] = acc

	refresh := "wrong-user-token"
	repo.refreshTokens[refresh] = &RefreshToken{
		AccountID: 999,
		Token:     refresh,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	_, _, err := service.RefreshTokens(context.Background(), refresh)
	assert.Error(t, err)
}

func TestLogoutRepoError(t *testing.T) {
	repo := &MockRepo{
		accounts: make(map[string]*Account),
		refreshTokens: map[string]*RefreshToken{
			"t": {AccountID: 1, Token: "t"},
		},
		deleteError: errors.New("db err"),
	}
	service := NewService(repo)
	err := service.Logout(context.Background(), 1)
	assert.Error(t, err)
}

func TestLoginInvalidHash(t *testing.T) {
	repo := &MockRepo{
		accounts: map[string]*Account{
			"bad@hash.com": {ID: 1, Email: "bad@hash.com", PasswordHash: "not-a-valid-hash"},
		},
		refreshTokens: make(map[string]*RefreshToken),
	}
	service := NewService(repo)
	_, _, err := service.Login(context.Background(), "bad@hash.com", "password")
	assert.Error(t, err)
}

func TestRegisterSavesRefreshToken(t *testing.T) {
	repo := &MockRepo{accounts: make(map[string]*Account), refreshTokens: make(map[string]*RefreshToken)}
	service := NewService(repo)
	_, _, refreshToken, err := service.Register(context.Background(), RegisterRequest{
		Email: "x@test.com", Password: "password", FirstName: "X", LastName: "Y",
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, repo.refreshTokens)
	assert.NotNil(t, repo.refreshTokens[refreshToken])
}

func TestUpdateProfileEmailConflict(t *testing.T) {
	repo := &MockRepo{
		accounts: map[string]*Account{
			"a@test.com": {ID: 1, Email: "a@test.com"},
			"b@test.com": {ID: 2, Email: "b@test.com"},
		},
		refreshTokens: make(map[string]*RefreshToken),
	}
	service := NewService(repo)

	_, err := service.UpdateProfile(context.Background(), 1, &UpdateProfileRequest{
		Email:     "b@test.com",
		FirstName: "A",
		LastName:  "B",
	})
	assert.Error(t, err)
}
