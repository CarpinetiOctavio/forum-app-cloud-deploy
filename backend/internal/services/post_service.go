package services

import (
	"errors"
	"strings"

	"tp06-testing/internal/models"
	"tp06-testing/internal/repository"
)

// Error message constants
const (
	ErrUserNotFound = "user not found"
	ErrPostNotFound = "post not found"
)

// PostService handles post and comment logic
type PostService struct {
	postRepo repository.PostRepository
	userRepo repository.UserRepository
}

// NewPostService creates a new instance
func NewPostService(postRepo repository.PostRepository, userRepo repository.UserRepository) *PostService {
	return &PostService{
		postRepo: postRepo,
		userRepo: userRepo,
	}
}

// CreatePost creates a new post
func (s *PostService) CreatePost(req *models.CreatePostRequest, userID int) (*models.Post, error) {
	if strings.TrimSpace(req.Title) == "" {
		return nil, errors.New("title is required")
	}

	if len(strings.TrimSpace(req.Title)) < 3 {
		return nil, errors.New("title must be at least 3 characters")
	}

	if strings.TrimSpace(req.Content) == "" {
		return nil, errors.New("content is required")
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New(ErrUserNotFound)
	}

	post := &models.Post{
		Title:   strings.TrimSpace(req.Title),
		Content: strings.TrimSpace(req.Content),
		UserID:  userID,
	}

	err = s.postRepo.Create(post)
	if err != nil {
		return nil, err
	}

	post.Username = user.Username

	return post, nil
}

// GetAllPosts retrieves all posts in the system.
// Returns an empty slice if there are no posts, never nil.
func (s *PostService) GetAllPosts() ([]*models.Post, error) {
	posts, err := s.postRepo.FindAll()
	if err != nil {
		return nil, err
	}

	// No posts: return empty slice (not an error)
	if posts == nil {
		return []*models.Post{}, nil
	}

	return posts, nil
}

// GetPostByID retrieves a specific post
func (s *PostService) GetPostByID(id int) (*models.Post, error) {
	if id <= 0 {
		return nil, errors.New("invalid id")
	}

	post, err := s.postRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if post == nil {
		return nil, errors.New(ErrPostNotFound)
	}

	return post, nil
}

// DeletePost removes a post (only the author can delete it)
func (s *PostService) DeletePost(postID int, userID int) error {
	post, err := s.postRepo.FindByID(postID)
	if err != nil {
		return err
	}
	if post == nil {
		return errors.New(ErrPostNotFound)
	}

	if post.UserID != userID {
		return errors.New("you do not have permission to delete this post")
	}

	return s.postRepo.Delete(postID)
}

// CreateComment adds a comment to a post
func (s *PostService) CreateComment(postID int, req *models.CreateCommentRequest, userID int) (*models.Comment, error) {
	if strings.TrimSpace(req.Content) == "" {
		return nil, errors.New("comment content is required")
	}

	post, err := s.postRepo.FindByID(postID)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, errors.New(ErrPostNotFound)
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New(ErrUserNotFound)
	}

	comment := &models.Comment{
		PostID:  postID,
		UserID:  userID,
		Content: strings.TrimSpace(req.Content),
	}

	err = s.postRepo.CreateComment(comment)
	if err != nil {
		return nil, err
	}

	comment.Username = user.Username

	return comment, nil
}

// GetCommentsByPostID retrieves all comments for a post
func (s *PostService) GetCommentsByPostID(postID int) ([]*models.Comment, error) {
	post, err := s.postRepo.FindByID(postID)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, errors.New(ErrPostNotFound)
	}

	comments, err := s.postRepo.FindCommentsByPostID(postID)
	if err != nil {
		return nil, err
	}

	if comments == nil {
		return []*models.Comment{}, nil
	}

	return comments, nil
}

func (s *PostService) DeleteComment(postID int, commentID int, userID int) error {
	post, err := s.postRepo.FindByID(postID)
	if err != nil {
		return err
	}
	if post == nil {
		return errors.New(ErrPostNotFound)
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New(ErrUserNotFound)
	}

	return s.postRepo.DeleteComment(postID, commentID, userID)
}
