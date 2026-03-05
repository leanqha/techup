package account

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

type MockRepo struct {
	accounts       map[string]*Account
	refreshTokens  map[string]*RefreshToken
	resetByHash    map[string]*PasswordResetToken
	resetByID      map[int]*PasswordResetToken
	nextResetID    int
	createErr      error
	saveError      error
	saveRefreshErr error
	getError       error
	deleteError    error
}

func (m *MockRepo) CreateAccount(_ context.Context, acc *Account) error {
	if m.createErr != nil {
		return m.createErr
	}
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
	for email, a := range m.accounts {
		if email == acc.Email && a.ID != acc.ID {
			return errors.New("email already exists")
		}
	}
	for email, a := range m.accounts {
		if a.ID == acc.ID {
			delete(m.accounts, email)
			break
		}
	}
	m.accounts[acc.Email] = acc
	return nil
}

func (m *MockRepo) ListAccounts(_ context.Context, _ AdminAccountsFilter) ([]Account, error) {
	result := make([]Account, 0, len(m.accounts))
	for _, acc := range m.accounts {
		result = append(result, *acc)
	}
	return result, nil
}

func (m *MockRepo) SaveRefreshToken(_ context.Context, token *RefreshToken) error {
	if m.saveRefreshErr != nil {
		return m.saveRefreshErr
	}
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

func (m *MockRepo) DeleteAccount(_ context.Context, userID int) error {
	if m.deleteError != nil {
		return m.deleteError
	}
	for token, rt := range m.refreshTokens {
		if rt.AccountID == userID {
			delete(m.refreshTokens, token)
		}
	}
	for email, acc := range m.accounts {
		if acc.ID == userID {
			delete(m.accounts, email)
			return nil
		}
	}
	return errors.New("account not found")
}

func (m *MockRepo) CreatePasswordResetToken(_ context.Context, token *PasswordResetToken) error {
	if m.saveError != nil {
		return m.saveError
	}
	if m.resetByHash == nil {
		m.resetByHash = make(map[string]*PasswordResetToken)
	}
	if m.resetByID == nil {
		m.resetByID = make(map[int]*PasswordResetToken)
	}
	m.nextResetID++
	token.ID = m.nextResetID
	token.CreatedAt = time.Now()
	m.resetByHash[token.TokenHash] = token
	m.resetByID[token.ID] = token
	return nil
}

func (m *MockRepo) GetPasswordResetTokenByHash(_ context.Context, tokenHash string) (*PasswordResetToken, error) {
	t, ok := m.resetByHash[tokenHash]
	if !ok {
		return nil, errors.New("not found")
	}
	return t, nil
}

func (m *MockRepo) MarkPasswordResetTokenUsed(_ context.Context, tokenID int) error {
	t, ok := m.resetByID[tokenID]
	if !ok {
		return errors.New("not found")
	}
	now := time.Now()
	t.UsedAt = &now
	return nil
}

func (m *MockRepo) DeletePasswordResetTokens(_ context.Context, accountID int) error {
	for hash, token := range m.resetByHash {
		if token.AccountID == accountID {
			delete(m.resetByHash, hash)
			delete(m.resetByID, token.ID)
		}
	}
	return nil
}

func TestRegisterAndLogin(t *testing.T) {
	_ = gofakeit.Seed(0)
	repo := &MockRepo{
		accounts:      make(map[string]*Account),
		refreshTokens: make(map[string]*RefreshToken),
	}

	service := NewService(repo)

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

	_, _, _, err = service.Register(context.Background(), RegisterRequest{
		Email:     "test@example.com",
		Password:  "password",
		FirstName: "John",
		LastName:  "Doe",
	})
	assert.Error(t, err)

	access, refresh, err := service.Login(context.Background(), "test@example.com", "password")
	assert.NoError(t, err)
	assert.NotEmpty(t, access)
	assert.NotEmpty(t, refresh)

	_, _, err = service.Login(context.Background(), "test@example.com", "wrongpassword")
	assert.Error(t, err)
}

func TestGetByID(t *testing.T) {
	_ = gofakeit.Seed(0)
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
	_ = gofakeit.Seed(0)
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

	err = service.repo.SaveRefreshToken(context.Background(), &RefreshToken{
		AccountID: acc.ID,
		Token:     refresh,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	})
	assert.NoError(t, err)
}

func TestLogout(t *testing.T) {
	_ = gofakeit.Seed(0)
	repo := &MockRepo{
		accounts:      make(map[string]*Account),
		refreshTokens: make(map[string]*RefreshToken),
	}

	service := NewService(repo)

	acc := &Account{ID: 1, Email: "user@example.com"}
	repo.accounts[acc.Email] = acc

	refreshToken := "refresh-token-123"
	repo.refreshTokens[refreshToken] = &RefreshToken{
		AccountID: acc.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	err := service.Logout(context.Background(), acc.ID, refreshToken)
	assert.NoError(t, err)

	_, exists := repo.refreshTokens[refreshToken]
	assert.False(t, exists)
}

func TestUpdateProfile(t *testing.T) {
	_ = gofakeit.Seed(0)
	repo := &MockRepo{
		accounts: make(map[string]*Account),
	}

	service := NewService(repo)

	acc := &Account{ID: 1, Email: "user@example.com", FirstName: "John", LastName: "Doe"}
	repo.accounts[acc.Email] = acc

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

	_, err = service.UpdateProfile(context.Background(), 999, &req)
	assert.Error(t, err)
}

func TestLoginRefreshLogoutEdgeCases(t *testing.T) {
	_ = gofakeit.Seed(0)
	repo := &MockRepo{
		accounts:      make(map[string]*Account),
		refreshTokens: make(map[string]*RefreshToken),
	}

	service := NewService(repo)

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

	access, refresh, err := service.Login(context.Background(), acc.Email, "password")
	assert.NoError(t, err)
	assert.NotEmpty(t, access)
	assert.NotEmpty(t, refresh)

	_, _, err = service.Login(context.Background(), acc.Email, "wrongpassword")
	assert.Error(t, err)

	_, _, err = service.Login(context.Background(), "missing@example.com", "password")
	assert.Error(t, err)

	err = repo.SaveRefreshToken(context.Background(), &RefreshToken{
		AccountID: acc.ID,
		Token:     refresh,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	})
	assert.NoError(t, err)

	_, _, err = service.RefreshTokens(context.Background(), "invalid-token")
	assert.Error(t, err)

	newAccess, newRefresh, err := service.RefreshTokens(context.Background(), refresh)
	assert.NoError(t, err)
	assert.NotEmpty(t, newAccess)
	assert.NotEmpty(t, newRefresh)

	err = service.Logout(context.Background(), acc.ID, refresh)
	assert.NoError(t, err)
	for _, rt := range repo.refreshTokens {
		assert.NotEqual(t, acc.ID, rt.AccountID)
	}

	err = service.Logout(context.Background(), 999, "")
	assert.NoError(t, err) // не должно падать, просто ничего не удаляет
}

func TestRegisterValidation(t *testing.T) {
	_ = gofakeit.Seed(0)
	repo := &MockRepo{accounts: make(map[string]*Account), refreshTokens: make(map[string]*RefreshToken)}
	service := NewService(repo)

	_, _, _, err := service.Register(context.Background(), RegisterRequest{
		Email: "", Password: "password", FirstName: "A", LastName: "B",
	})
	assert.Error(t, err)

	_, _, _, err = service.Register(context.Background(), RegisterRequest{
		Email: "a@b.c", Password: "123", FirstName: "A", LastName: "B",
	})
	assert.Error(t, err)

	_, _, _, err = service.Register(context.Background(), RegisterRequest{
		Email: "a@b.c", Password: "password", FirstName: "", LastName: "B",
	})
	assert.Error(t, err)

	_, _, _, err = service.Register(context.Background(), RegisterRequest{
		Email: "a@b.c", Password: "password", FirstName: "A", LastName: "",
	})
	assert.Error(t, err)
}

func TestPasswordHashing(t *testing.T) {
	_ = gofakeit.Seed(0)
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
	_ = gofakeit.Seed(0)
	repo := &MockRepo{accounts: make(map[string]*Account), refreshTokens: make(map[string]*RefreshToken)}
	service := NewService(repo)
	acc := &Account{ID: 1, Email: gofakeit.Email()}
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
	_ = gofakeit.Seed(0)
	repo := &MockRepo{accounts: make(map[string]*Account), refreshTokens: make(map[string]*RefreshToken)}
	service := NewService(repo)

	acc := &Account{ID: 1, Email: gofakeit.Email()}
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
	_ = gofakeit.Seed(0)
	repo := &MockRepo{
		accounts: make(map[string]*Account),
		refreshTokens: map[string]*RefreshToken{
			"t": {AccountID: 1, Token: "t"},
		},
		deleteError: errors.New("db err"),
	}
	service := NewService(repo)
	err := service.Logout(context.Background(), 1, "")
	assert.Error(t, err)
}

func TestLoginInvalidHash(t *testing.T) {
	_ = gofakeit.Seed(0)
	repo := &MockRepo{
		accounts: map[string]*Account{
			gofakeit.Email(): {ID: 1, Email: gofakeit.Email(), PasswordHash: "not-a-valid-hash"},
		},
		refreshTokens: make(map[string]*RefreshToken),
	}
	service := NewService(repo)
	var email string
	for k := range repo.accounts {
		email = k
		break
	}
	_, _, err := service.Login(context.Background(), email, gofakeit.Password(true, true, true, false, false, 12))
	assert.Error(t, err)
}

func TestRegisterSavesRefreshToken(t *testing.T) {
	repo := &MockRepo{accounts: make(map[string]*Account), refreshTokens: make(map[string]*RefreshToken)}
	service := NewService(repo)

	acc, _, _, err := service.Register(context.Background(), RegisterRequest{
		Email:     "tokens@test.com",
		Password:  "password",
		FirstName: "A",
		LastName:  "B",
	})
	assert.NoError(t, err)
	assert.NotZero(t, acc.ID)

	assert.NotEmpty(t, repo.refreshTokens)
	for _, rt := range repo.refreshTokens {
		assert.Equal(t, acc.ID, rt.AccountID)
		assert.NotEmpty(t, rt.Token)
		break
	}
}

func TestRegisterCreateAccountError(t *testing.T) {
	repo := &MockRepo{
		accounts:      make(map[string]*Account),
		refreshTokens: make(map[string]*RefreshToken),
		createErr:     errors.New("create failed"),
	}
	service := NewService(repo)

	_, _, _, err := service.Register(context.Background(), RegisterRequest{
		Email:     "create@test.com",
		Password:  "password",
		FirstName: "A",
		LastName:  "B",
	})
	assert.Error(t, err)
}

func TestChangePasswordSuccess(t *testing.T) {
	repo := &MockRepo{
		accounts:      make(map[string]*Account),
		refreshTokens: make(map[string]*RefreshToken),
	}
	service := NewService(repo)

	hash, _ := bcrypt.GenerateFromPassword([]byte("oldpass"), bcrypt.DefaultCost)
	acc := &Account{ID: 1, Email: "user@example.com", PasswordHash: string(hash)}
	repo.accounts[acc.Email] = acc

	err := service.ChangePassword(context.Background(), acc.ID, &ChangePasswordRequest{
		OldPassword: "oldpass",
		NewPassword: "newpass123",
	})
	assert.NoError(t, err)

	updated, _ := repo.GetByID(context.Background(), acc.ID)
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(updated.PasswordHash), []byte("newpass123")))
}

func TestChangePasswordWrongOldPassword(t *testing.T) {
	repo := &MockRepo{
		accounts:      make(map[string]*Account),
		refreshTokens: make(map[string]*RefreshToken),
	}
	service := NewService(repo)

	hash, _ := bcrypt.GenerateFromPassword([]byte("oldpass"), bcrypt.DefaultCost)
	acc := &Account{ID: 1, Email: "user@example.com", PasswordHash: string(hash)}
	repo.accounts[acc.Email] = acc

	err := service.ChangePassword(context.Background(), acc.ID, &ChangePasswordRequest{
		OldPassword: "wrong",
		NewPassword: "newpass123",
	})
	assert.Error(t, err)
}

func TestSetRoleAdminAndForbidden(t *testing.T) {
	repo := &MockRepo{accounts: make(map[string]*Account), refreshTokens: make(map[string]*RefreshToken)}
	service := NewService(repo)

	admin := &Account{ID: 1, Email: "admin@example.com", Role: "admin"}
	user := &Account{ID: 2, Email: "user@example.com", Role: "student"}
	repo.accounts[admin.Email] = admin
	repo.accounts[user.Email] = user

	err := service.SetRole(context.Background(), admin.ID, &SetRoleRequest{UserID: user.ID, Role: "teacher"})
	assert.NoError(t, err)
	updated, _ := repo.GetByID(context.Background(), user.ID)
	assert.Equal(t, "teacher", updated.Role)

	nonAdmin := &Account{ID: 3, Email: "staff@example.com", Role: "staff"}
	repo.accounts[nonAdmin.Email] = nonAdmin
	err = service.SetRole(context.Background(), nonAdmin.ID, &SetRoleRequest{UserID: user.ID, Role: "admin"})
	assert.Error(t, err)
}

func TestRefreshTokenInvalidType(t *testing.T) {
	repo := &MockRepo{accounts: make(map[string]*Account), refreshTokens: make(map[string]*RefreshToken)}
	service := NewService(repo)

	acc := &Account{ID: 1, Email: "user@example.com"}
	repo.accounts[acc.Email] = acc

	access, _, err := GenerateTokens(acc)
	assert.NoError(t, err)

	_, _, err = service.RefreshTokens(context.Background(), access)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid token type")
}

func TestLoginSaveRefreshTokenError(t *testing.T) {
	repo := &MockRepo{
		accounts:       make(map[string]*Account),
		refreshTokens:  make(map[string]*RefreshToken),
		saveRefreshErr: errors.New("save refresh failed"),
	}
	service := NewService(repo)

	hash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	acc := &Account{ID: 1, Email: "user@example.com", PasswordHash: string(hash)}
	repo.accounts[acc.Email] = acc

	_, _, err := service.Login(context.Background(), acc.Email, "password")
	assert.Error(t, err)
}

func TestRegisterSaveRefreshTokenError(t *testing.T) {
	repo := &MockRepo{
		accounts:       make(map[string]*Account),
		refreshTokens:  make(map[string]*RefreshToken),
		saveRefreshErr: errors.New("save refresh failed"),
	}
	service := NewService(repo)

	_, _, _, err := service.Register(context.Background(), RegisterRequest{
		Email:     "test@example.com",
		Password:  "password",
		FirstName: "John",
		LastName:  "Doe",
	})
	assert.Error(t, err)
}

func TestDeleteAccountError(t *testing.T) {
	repo := &MockRepo{
		accounts:      make(map[string]*Account),
		refreshTokens: make(map[string]*RefreshToken),
		deleteError:   errors.New("delete failed"),
	}
	service := NewService(repo)

	err := service.DeleteAccount(context.Background(), 1)
	assert.Error(t, err)
}
