package mocks

import (
	"forum-app-qa-pipeline/internal/models"

	"github.com/stretchr/testify/mock"
)

// MockPostRepository is a mock of PostRepository
type MockPostRepository struct {
	mock.Mock
}

// Create simulates creating a post
func (m *MockPostRepository) Create(post *models.Post) error {
	args := m.Called(post)
	return args.Error(0)
}

// FindAll simulates retrieving all posts
func (m *MockPostRepository) FindAll() ([]*models.Post, error) {
	args := m.Called()

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]*models.Post), args.Error(1)
}

// FindByID simulates looking up a post by ID
func (m *MockPostRepository) FindByID(id int) (*models.Post, error) {
	args := m.Called(id)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*models.Post), args.Error(1)
}

// Update simulates updating a post
func (m *MockPostRepository) Update(post *models.Post) error {
	args := m.Called(post)
	return args.Error(0)
}

// Delete simulates deleting a post
func (m *MockPostRepository) Delete(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

// CreateComment simulates creating a comment
func (m *MockPostRepository) CreateComment(comment *models.Comment) error {
	args := m.Called(comment)
	return args.Error(0)
}

// FindCommentsByPostID simulates retrieving comments for a post
func (m *MockPostRepository) FindCommentsByPostID(postID int) ([]*models.Comment, error) {
	args := m.Called(postID)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]*models.Comment), args.Error(1)
}

// FindCommentByID simulates looking up a single comment by ID
func (m *MockPostRepository) FindCommentByID(commentID int) (*models.Comment, error) {
	args := m.Called(commentID)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*models.Comment), args.Error(1)
}

// UpdateComment simulates updating a comment
func (m *MockPostRepository) UpdateComment(comment *models.Comment) error {
	args := m.Called(comment)
	return args.Error(0)
}

// DeleteComment simulates deleting a comment
func (m *MockPostRepository) DeleteComment(postID int, commentID int, userID int) error {
	args := m.Called(postID, commentID, userID)
	return args.Error(0)
}
