package services

import (
	"strings"
	"testing"

	"forum-app-qa-pipeline/internal/models"
	"forum-app-qa-pipeline/internal/services"
	"forum-app-qa-pipeline/tests/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// TestRegister_Success verifies a user registers successfully, and that the
// password reaching the repository is bcrypt-hashed, never the plaintext
// input.
//
// Unlike TestLogin_Success/TestLogin_IncorrectPassword, this test cannot use
// bcrypt.MinCost: it calls the real AuthService.Register, which hashes with
// bcrypt.DefaultCost internally (see ADR-008) - the cost isn't a test-side
// choice here, it's dictated by production code. That makes this the one
// test in this file that pays bcrypt's real ~60-100ms cost. That's expected
// for this single test, not a regression - mocking the hashing itself to
// avoid it would test nothing.
func TestRegister_Success(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockUserRepository)
	authService := services.NewAuthService(mockRepo)

	// Configure the mock: the email does not exist yet
	mockRepo.On("FindByEmail", "test@example.com").Return(nil, nil)

	// Configure the mock: Create should execute successfully, capturing the
	// user it was called with so the ASSERT step can inspect the password
	var createdUser *models.User
	mockRepo.On("Create", mock.AnythingOfType("*models.User")).
		Run(func(args mock.Arguments) {
			createdUser = args.Get(0).(*models.User)
		}).
		Return(nil)

	req := &models.RegisterRequest{
		Email:    "test@example.com",
		Password: "123456",
		Username: "testuser",
	}

	// ACT
	user, err := authService.Register(req)

	// ASSERT
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "testuser", user.Username)

	// The password persisted to the repository must be hashed, not plaintext
	assert.NotEqual(t, "123456", createdUser.Password)
	hashErr := bcrypt.CompareHashAndPassword([]byte(createdUser.Password), []byte("123456"))
	assert.NoError(t, hashErr)

	// Verify the mock's methods were called
	mockRepo.AssertExpectations(t)
}

// TestRegister_EmptyEmail verifies registration fails when the email is empty
func TestRegister_EmptyEmail(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockUserRepository)
	authService := services.NewAuthService(mockRepo)

	req := &models.RegisterRequest{
		Email:    "", // Empty email
		Password: "123456",
		Username: "testuser",
	}

	// ACT
	user, err := authService.Register(req)

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, "email is required", err.Error())

	// Should NOT have called the DB because validation failed first
	mockRepo.AssertNotCalled(t, "FindByEmail")
	mockRepo.AssertNotCalled(t, "Create")
}

// TestRegister_InvalidEmail verifies registration fails when the email is missing the @ symbol
func TestRegister_InvalidEmail(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockUserRepository)
	authService := services.NewAuthService(mockRepo)

	req := &models.RegisterRequest{
		Email:    "invalidemail", // Missing @
		Password: "123456",
		Username: "testuser",
	}

	// ACT
	user, err := authService.Register(req)

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, "email must be valid", err.Error())
}

// TestRegister_PasswordTooShort verifies registration fails when the password is under 6 characters
func TestRegister_PasswordTooShort(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockUserRepository)
	authService := services.NewAuthService(mockRepo)

	req := &models.RegisterRequest{
		Email:    "test@example.com",
		Password: "123", // Too short
		Username: "testuser",
	}

	// ACT
	user, err := authService.Register(req)

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, "password must be at least 6 characters", err.Error())
}

// TestRegister_PasswordTooLong verifies registration fails when the password exceeds bcrypt's 72-byte input limit
func TestRegister_PasswordTooLong(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockUserRepository)
	authService := services.NewAuthService(mockRepo)

	req := &models.RegisterRequest{
		Email:    "test@example.com",
		Password: strings.Repeat("a", 73), // One byte over bcrypt's 72-byte limit
		Username: "testuser",
	}

	// ACT
	user, err := authService.Register(req)

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, "password must not exceed 72 characters", err.Error())

	// Should NOT have called the DB because validation failed first
	mockRepo.AssertNotCalled(t, "FindByEmail")
	mockRepo.AssertNotCalled(t, "Create")
}

// TestRegister_EmptyUsername verifies registration fails when the username is empty
func TestRegister_EmptyUsername(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockUserRepository)
	authService := services.NewAuthService(mockRepo)

	req := &models.RegisterRequest{
		Email:    "test@example.com",
		Password: "123456",
		Username: "", // Empty username
	}

	// ACT
	user, err := authService.Register(req)

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, "username is required", err.Error())
}

// TestRegister_DuplicateEmail verifies registration fails when the email is already registered
func TestRegister_DuplicateEmail(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockUserRepository)
	authService := services.NewAuthService(mockRepo)

	existingUser := &models.User{
		ID:       1,
		Email:    "test@example.com",
		Username: "existinguser",
	}

	// Configure the mock: the email already exists
	mockRepo.On("FindByEmail", "test@example.com").Return(existingUser, nil)

	req := &models.RegisterRequest{
		Email:    "test@example.com",
		Password: "123456",
		Username: "testuser",
	}

	// ACT
	user, err := authService.Register(req)

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, "email is already registered", err.Error())

	// Should NOT call Create because the email already exists
	mockRepo.AssertNotCalled(t, "Create")
}

// TestLogin_Success verifies a user logs in successfully
func TestLogin_Success(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockUserRepository)
	authService := services.NewAuthService(mockRepo)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.MinCost)
	assert.NoError(t, err)

	existingUser := &models.User{
		ID:       1,
		Email:    "test@example.com",
		Password: string(hashedPassword),
		Username: "testuser",
	}

	// Configure the mock: the user exists
	mockRepo.On("FindByEmail", "test@example.com").Return(existingUser, nil)

	creds := &models.Credentials{
		Email:    "test@example.com",
		Password: "123456",
	}

	// ACT
	user, loginErr := authService.Login(creds)

	// ASSERT
	assert.NoError(t, loginErr)
	assert.NotNil(t, user)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "testuser", user.Username)

	mockRepo.AssertExpectations(t)
}

// TestLogin_EmptyEmail verifies login fails when the email is empty
func TestLogin_EmptyEmail(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockUserRepository)
	authService := services.NewAuthService(mockRepo)

	creds := &models.Credentials{
		Email:    "",
		Password: "123456",
	}

	// ACT
	user, err := authService.Login(creds)

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, "email is required", err.Error())

	mockRepo.AssertNotCalled(t, "FindByEmail")
}

// TestLogin_EmptyPassword verifies login fails when the password is empty
func TestLogin_EmptyPassword(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockUserRepository)
	authService := services.NewAuthService(mockRepo)

	creds := &models.Credentials{
		Email:    "test@example.com",
		Password: "",
	}

	// ACT
	user, err := authService.Login(creds)

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, "password is required", err.Error())
}

// TestLogin_UserNotFound verifies login fails when no user matches the given email
func TestLogin_UserNotFound(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockUserRepository)
	authService := services.NewAuthService(mockRepo)

	// Configure the mock: the user does NOT exist
	mockRepo.On("FindByEmail", "noexiste@example.com").Return(nil, nil)

	creds := &models.Credentials{
		Email:    "noexiste@example.com",
		Password: "123456",
	}

	// ACT
	user, err := authService.Login(creds)

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, "invalid credentials", err.Error())

	mockRepo.AssertExpectations(t)
}

// TestLogin_IncorrectPassword verifies login fails when the password does not match
func TestLogin_IncorrectPassword(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockUserRepository)
	authService := services.NewAuthService(mockRepo)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.MinCost)
	assert.NoError(t, err)

	existingUser := &models.User{
		ID:       1,
		Email:    "test@example.com",
		Password: string(hashedPassword),
		Username: "testuser",
	}

	mockRepo.On("FindByEmail", "test@example.com").Return(existingUser, nil)

	creds := &models.Credentials{
		Email:    "test@example.com",
		Password: "wrongpassword", // Incorrect password
	}

	// ACT
	user, err := authService.Login(creds)

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, "invalid credentials", err.Error())

	mockRepo.AssertExpectations(t)
}
