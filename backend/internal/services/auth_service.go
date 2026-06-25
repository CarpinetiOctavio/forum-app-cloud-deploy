package services

import (
	"errors"
	"strings"

	"tp06-testing/internal/models"
	"tp06-testing/internal/repository"
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

// Register registers a new user and validates business rules
func (s *AuthService) Register(req *models.RegisterRequest) (*models.User, error) {
	// Validation 1: email must not be empty
	if strings.TrimSpace(req.Email) == "" {
		return nil, errors.New("email is required")
	}

	// Validation 2: email must contain @
	if !strings.Contains(req.Email, "@") {
		return nil, errors.New("email must be valid")
	}

	// Validation 3: password must be at least 6 characters
	if len(req.Password) < 6 {
		return nil, errors.New("password must be at least 6 characters")
	}

	// Validation 4: username must not be empty
	if strings.TrimSpace(req.Username) == "" {
		return nil, errors.New("username is required")
	}

	// Validation 5: verify email is not already registered
	existingUser, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("email is already registered")
	}

	// Create the user
	user := &models.User{
		Email:    strings.ToLower(strings.TrimSpace(req.Email)),
		Password: req.Password, // in production: hash with bcrypt
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
	// Validation 1: email must not be empty
	if strings.TrimSpace(creds.Email) == "" {
		return nil, errors.New("email is required")
	}

	// Validation 2: password must not be empty
	if creds.Password == "" {
		return nil, errors.New("password is required")
	}

	// Find user by email
	user, err := s.userRepo.FindByEmail(strings.ToLower(strings.TrimSpace(creds.Email)))
	if err != nil {
		return nil, err
	}

	// Validation 3: user must exist
	if user == nil {
		return nil, errors.New("invalid credentials")
	}

	// Validation 4: password must match
	// in production: use bcrypt.CompareHashAndPassword
	if user.Password != creds.Password {
		return nil, errors.New("invalid credentials")
	}

	return user, nil
}
