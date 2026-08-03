package mocks

import (
	"forum-app-qa-pipeline/internal/models"

	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock of UserRepository
type MockUserRepository struct {
	mock.Mock
}

// Create simulates creating a user
func (m *MockUserRepository) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

// FindByEmail simulates looking up a user by email
func (m *MockUserRepository) FindByEmail(email string) (*models.User, error) {
	args := m.Called(email)

	// If configured to return nil (user not found)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*models.User), args.Error(1)
}

// FindByID simulates looking up a user by ID
func (m *MockUserRepository) FindByID(id int) (*models.User, error) {
	args := m.Called(id)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*models.User), args.Error(1)
}
