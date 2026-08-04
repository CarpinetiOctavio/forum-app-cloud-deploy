package services

import (
	"errors"
	"strings"

	"forum-app-cloud-deploy/internal/models"
	"forum-app-cloud-deploy/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

// AuthService handles authentication logic
type AuthService struct {
	userRepo repository.UserRepository
}

// NewAuthService creates a new instance
func NewAuthService(userRepo repository.UserRepository) *AuthService {
	return &AuthService{
		userRepo: userRepo,
	}
}

// Register registers a new user
// Business rules are validated here
func (s *AuthService) Register(req *models.RegisterRequest) (*models.User, error) {
	// Validation 1: Email cannot be empty
	if strings.TrimSpace(req.Email) == "" {
		return nil, errors.New("email is required")
	}

	// Validation 2: Email must contain @
	if !strings.Contains(req.Email, "@") {
		return nil, errors.New("email must be valid")
	}

	// Validation 3: Password must be at least 6 characters
	if len(req.Password) < 6 {
		return nil, errors.New("password must be at least 6 characters")
	}

	// Validation 4: Password must not exceed 72 characters (bcrypt's own input limit)
	if len(req.Password) > 72 {
		return nil, errors.New("password must not exceed 72 characters")
	}

	// Validation 5: Username cannot be empty
	if strings.TrimSpace(req.Username) == "" {
		return nil, errors.New("username is required")
	}

	// Validation 6: Verify the email is not already registered
	existingUser, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("email is already registered")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create the user
	user := &models.User{
		Email:    strings.ToLower(strings.TrimSpace(req.Email)),
		Password: string(hashedPassword),
		Username: strings.TrimSpace(req.Username),
	}

	err = s.userRepo.Create(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// Login authenticates a user
func (s *AuthService) Login(creds *models.Credentials) (*models.User, error) {
	// Validation 1: Email cannot be empty
	if strings.TrimSpace(creds.Email) == "" {
		return nil, errors.New("email is required")
	}

	// Validation 2: Password cannot be empty
	if creds.Password == "" {
		return nil, errors.New("password is required")
	}

	// Look up the user by email
	user, err := s.userRepo.FindByEmail(strings.ToLower(strings.TrimSpace(creds.Email)))
	if err != nil {
		return nil, err
	}

	// Validation 3: User must exist
	if user == nil {
		return nil, errors.New("invalid credentials")
	}

	// Validation 4: Password must match
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	return user, nil
}
