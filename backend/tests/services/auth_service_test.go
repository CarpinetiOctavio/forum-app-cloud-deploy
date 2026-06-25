package services

import (
	"testing"

	"tp06-testing/internal/models"
	"tp06-testing/internal/services"
	"tp06-testing/tests/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Common test constants
const (
	testEmail    = "test@example.com"
	testPassword = "123456"
	testUsername = "testuser"
)

// TestRegister_Success tests successful user registration
func TestRegister_Success(t *testing.T) {
	// ARRANGE: set up mock and test data
	mockRepo := new(mocks.MockUserRepository)
	authService := services.NewAuthService(mockRepo)

	// Configure mock: email does NOT exist (returns nil)
	mockRepo.On("FindByEmail", testEmail).Return(nil, nil)

	// Configure mock: Create should succeed
	mockRepo.On("Create", mock.AnythingOfType("*models.User")).Return(nil)

	req := &models.RegisterRequest{
		Email:    testEmail,
		Password: testPassword,
		Username: testUsername,
	}

	// ACT: execute the function under test
	user, err := authService.Register(req)

	// ASSERT: verify results
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, testEmail, user.Email)
	assert.Equal(t, testUsername, user.Username)

	// Verify mock methods were called
	mockRepo.AssertExpectations(t)
}

// TestRegister_EmailVacio tests failure with empty email
func TestRegister_EmailVacio(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockUserRepository)
	authService := services.NewAuthService(mockRepo)

	req := &models.RegisterRequest{
		Email:    "", // empty email
		Password: testPassword,
		Username: testUsername,
	}

	// ACT
	user, err := authService.Register(req)

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, "email is required", err.Error())

	// Must not call the DB because validation failed first
	mockRepo.AssertNotCalled(t, "FindByEmail")
	mockRepo.AssertNotCalled(t, "Create")
}

// TestRegister_EmailInvalido tests failure with email missing @
func TestRegister_EmailInvalido(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockUserRepository)
	authService := services.NewAuthService(mockRepo)

	req := &models.RegisterRequest{
		Email:    "invalidemail", // missing @
		Password: testPassword,
		Username: testUsername,
	}

	// ACT
	user, err := authService.Register(req)

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, "email must be valid", err.Error())
}

// TestRegister_PasswordCorto tests failure with password shorter than 6 characters
func TestRegister_PasswordCorto(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockUserRepository)
	authService := services.NewAuthService(mockRepo)

	req := &models.RegisterRequest{
		Email:    testEmail,
		Password: "123", // too short
		Username: testUsername,
	}

	// ACT
	user, err := authService.Register(req)

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, "password must be at least 6 characters", err.Error())
}

// TestRegister_UsernameVacio tests failure with empty username
func TestRegister_UsernameVacio(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockUserRepository)
	authService := services.NewAuthService(mockRepo)

	req := &models.RegisterRequest{
		Email:    testEmail,
		Password: testPassword,
		Username: "", // empty username
	}

	// ACT
	user, err := authService.Register(req)

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, "username is required", err.Error())
}

// TestRegister_EmailDuplicado tests failure when email already exists
func TestRegister_EmailDuplicado(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockUserRepository)
	authService := services.NewAuthService(mockRepo)

	existingUser := &models.User{
		ID:       1,
		Email:    testEmail,
		Username: "existinguser",
	}

	// Configure mock: email ALREADY exists
	mockRepo.On("FindByEmail", testEmail).Return(existingUser, nil)

	req := &models.RegisterRequest{
		Email:    testEmail,
		Password: testPassword,
		Username: testUsername,
	}

	// ACT
	user, err := authService.Register(req)

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, "email is already registered", err.Error())

	// Must not call Create because email already exists
	mockRepo.AssertNotCalled(t, "Create")
}

// TestLogin_Success tests successful login
func TestLogin_Success(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockUserRepository)
	authService := services.NewAuthService(mockRepo)

	existingUser := &models.User{
		ID:       1,
		Email:    testEmail,
		Password: testPassword,
		Username: testUsername,
	}

	// Configure mock: user exists
	mockRepo.On("FindByEmail", testEmail).Return(existingUser, nil)

	creds := &models.Credentials{
		Email:    testEmail,
		Password: testPassword,
	}

	// ACT
	user, err := authService.Login(creds)

	// ASSERT
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, testEmail, user.Email)
	assert.Equal(t, testUsername, user.Username)

	mockRepo.AssertExpectations(t)
}

// TestLogin_EmailVacio tests failure with empty email
func TestLogin_EmailVacio(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockUserRepository)
	authService := services.NewAuthService(mockRepo)

	creds := &models.Credentials{
		Email:    "",
		Password: testPassword,
	}

	// ACT
	user, err := authService.Login(creds)

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, "email is required", err.Error())

	mockRepo.AssertNotCalled(t, "FindByEmail")
}

// TestLogin_PasswordVacio tests failure with empty password
func TestLogin_PasswordVacio(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockUserRepository)
	authService := services.NewAuthService(mockRepo)

	creds := &models.Credentials{
		Email:    testEmail,
		Password: "",
	}

	// ACT
	user, err := authService.Login(creds)

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, "password is required", err.Error())
}

// TestLogin_UsuarioNoExiste tests failure when user does not exist
func TestLogin_UsuarioNoExiste(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockUserRepository)
	authService := services.NewAuthService(mockRepo)

	// Configure mock: user does NOT exist
	mockRepo.On("FindByEmail", "noexiste@example.com").Return(nil, nil)

	creds := &models.Credentials{
		Email:    "noexiste@example.com",
		Password: testPassword,
	}

	// ACT
	user, err := authService.Login(creds)

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, "invalid credentials", err.Error())

	mockRepo.AssertExpectations(t)
}

// TestLogin_PasswordIncorrecta tests failure with incorrect password
func TestLogin_PasswordIncorrecta(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockUserRepository)
	authService := services.NewAuthService(mockRepo)

	existingUser := &models.User{
		ID:       1,
		Email:    testEmail,
		Password: testPassword,
		Username: testUsername,
	}

	mockRepo.On("FindByEmail", testEmail).Return(existingUser, nil)

	creds := &models.Credentials{
		Email:    testEmail,
		Password: "wrongpassword", // wrong password
	}

	// ACT
	user, err := authService.Login(creds)

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, "invalid credentials", err.Error())

	mockRepo.AssertExpectations(t)
}
