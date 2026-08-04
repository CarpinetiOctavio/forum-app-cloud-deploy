package services

import (
	"errors"
	"strings"

	"forum-app-cloud-deploy/internal/models"
	"forum-app-cloud-deploy/internal/repository"
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
	// Validation 1: Title cannot be empty
	if strings.TrimSpace(req.Title) == "" {
		return nil, errors.New("title is required")
	}

	// Validation 2: Title must be at least 3 characters
	if len(strings.TrimSpace(req.Title)) < 3 {
		return nil, errors.New("title must be at least 3 characters")
	}

	// Validation 3: Content cannot be empty
	if strings.TrimSpace(req.Content) == "" {
		return nil, errors.New("content is required")
	}

	// Validation 4: User must exist
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Create the post
	post := &models.Post{
		Title:   strings.TrimSpace(req.Title),
		Content: strings.TrimSpace(req.Content),
		UserID:  userID,
	}

	err = s.postRepo.Create(post)
	if err != nil {
		return nil, err
	}

	// Add the username for the response
	post.Username = user.Username

	return post, nil
}

// GetAllPosts retrieves all posts
func (s *PostService) GetAllPosts() ([]*models.Post, error) {
	posts, err := s.postRepo.FindAll()
	if err != nil {
		return nil, err
	}

	// If there are no posts, return an empty list (not an error)
	if posts == nil {
		return []*models.Post{}, nil
	}

	return posts, nil
}

// GetPostByID retrieves a specific post
func (s *PostService) GetPostByID(id int) (*models.Post, error) {
	// Validation: ID must be positive
	if id <= 0 {
		return nil, errors.New("invalid id")
	}

	post, err := s.postRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if post == nil {
		return nil, errors.New("post not found")
	}

	return post, nil
}

// EditPost updates an existing post's title and content (only the author may do so)
func (s *PostService) EditPost(postID int, req *models.EditPostRequest, userID int) (*models.Post, error) {
	// Validation 1: Title cannot be empty
	if strings.TrimSpace(req.Title) == "" {
		return nil, errors.New("title is required")
	}

	// Validation 2: Title must be at least 3 characters
	if len(strings.TrimSpace(req.Title)) < 3 {
		return nil, errors.New("title must be at least 3 characters")
	}

	// Validation 3: Content cannot be empty
	if strings.TrimSpace(req.Content) == "" {
		return nil, errors.New("content is required")
	}

	// Validation 4: Post must exist
	post, err := s.postRepo.FindByID(postID)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, errors.New("post not found")
	}

	// Validation 5: Only the author may edit
	if post.UserID != userID {
		return nil, errors.New("you do not have permission to edit this post")
	}

	// Apply the edit
	post.Title = strings.TrimSpace(req.Title)
	post.Content = strings.TrimSpace(req.Content)

	if err := s.postRepo.Update(post); err != nil {
		return nil, err
	}

	return post, nil
}

// DeletePost removes a post (only the author may do so)
func (s *PostService) DeletePost(postID int, userID int) error {
	// Validation 1: Post must exist
	post, err := s.postRepo.FindByID(postID)
	if err != nil {
		return err
	}
	if post == nil {
		return errors.New("post not found")
	}

	// Validation 2: Only the author may delete
	if post.UserID != userID {
		return errors.New("you do not have permission to delete this post")
	}

	// Delete
	return s.postRepo.Delete(postID)
}

// CreateComment adds a comment to a post
func (s *PostService) CreateComment(postID int, req *models.CreateCommentRequest, userID int) (*models.Comment, error) {
	// Validation 1: Content cannot be empty
	if strings.TrimSpace(req.Content) == "" {
		return nil, errors.New("comment content is required")
	}

	// Validation 2: Post must exist
	post, err := s.postRepo.FindByID(postID)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, errors.New("post not found")
	}

	// Validation 3: User must exist
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Create the comment
	comment := &models.Comment{
		PostID:  postID,
		UserID:  userID,
		Content: strings.TrimSpace(req.Content),
	}

	err = s.postRepo.CreateComment(comment)
	if err != nil {
		return nil, err
	}

	// Add the username for the response
	comment.Username = user.Username

	return comment, nil
}

// GetCommentsByPostID retrieves all comments for a post
func (s *PostService) GetCommentsByPostID(postID int) ([]*models.Comment, error) {
	// Validation: Post must exist
	post, err := s.postRepo.FindByID(postID)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, errors.New("post not found")
	}

	comments, err := s.postRepo.FindCommentsByPostID(postID)
	if err != nil {
		return nil, err
	}

	// If there are no comments, return an empty list
	if comments == nil {
		return []*models.Comment{}, nil
	}

	return comments, nil
}

// EditComment updates an existing comment's content (only the author may do so)
func (s *PostService) EditComment(postID int, commentID int, req *models.EditCommentRequest, userID int) (*models.Comment, error) {
	// Validation 1: Content cannot be empty
	if strings.TrimSpace(req.Content) == "" {
		return nil, errors.New("comment content is required")
	}

	// Validation 2: Post must exist
	post, err := s.postRepo.FindByID(postID)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, errors.New("post not found")
	}

	// Validation 3: Comment must exist and belong to this post
	comment, err := s.postRepo.FindCommentByID(commentID)
	if err != nil {
		return nil, err
	}
	if comment == nil || comment.PostID != postID {
		return nil, errors.New("comment not found")
	}

	// Validation 4: Only the author may edit
	if comment.UserID != userID {
		return nil, errors.New("you do not have permission to edit this comment")
	}

	// Apply the edit
	comment.Content = strings.TrimSpace(req.Content)

	if err := s.postRepo.UpdateComment(comment); err != nil {
		return nil, err
	}

	return comment, nil
}

func (s *PostService) DeleteComment(postID int, commentID int, userID int) error {
	// Validation: Post must exist
	post, err := s.postRepo.FindByID(postID)
	if err != nil {
		return err
	}
	if post == nil {
		return errors.New("post not found")
	}

	// Validation: User must exist
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	// Delete comment (only the author may do so)
	return s.postRepo.DeleteComment(postID, commentID, userID)
}
