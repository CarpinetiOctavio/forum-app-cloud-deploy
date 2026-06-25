package services

import (
	"errors"
	"testing"

	"forum-app-cloud-deploy/internal/models"
	"forum-app-cloud-deploy/internal/services"
	"forum-app-cloud-deploy/tests/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestCreatePost_Success tests successful post creation
func TestCreatePost_Success(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	// ← ADD THIS
	existingUser := &models.User{
		ID:       1,
		Email:    "test@example.com",
		Username: "testuser",
	}
	mockUserRepo.On("FindByID", 1).Return(existingUser, nil)
	// ← END

	// Configure mock: Create should succeed
	mockRepo.On("Create", mock.AnythingOfType("*models.Post")).Return(nil)

	req := &models.CreatePostRequest{
		Title:   "Test Post",
		Content: "This is a test post",
	}

	// ACT
	post, err := postService.CreatePost(req, 1)

	// ASSERT
	assert.NoError(t, err)
	assert.NotNil(t, post)
	assert.Equal(t, "Test Post", post.Title)
	assert.Equal(t, "This is a test post", post.Content)

	// Verify mock methods were called
	mockRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t) // ← ADD THIS TOO
}

// TestCreatePost_UserNotFound: el userId no existe -> error
func TestCreatePost_UserNotFound(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	// Configure mock: user FindByID returns nil (does not exist)
	mockUserRepo.On("FindByID", 999).Return(nil, nil)

	req := &models.CreatePostRequest{
		Title:   "Test Post",
		Content: "This is a test post",
	}

	// ACT
	post, err := postService.CreatePost(req, 999)

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, post)
	assert.Equal(t, "user not found", err.Error())

	mockUserRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "Create")
}

// TestCreatePost_RepoError: el repositorio falla al crear -> se propaga error
func TestCreatePost_RepoError(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	// User exists
	existingUser := &models.User{ID: 1, Email: "u@u.com", Username: "u"}
	mockUserRepo.On("FindByID", 1).Return(existingUser, nil)

	// Repo Create fails
	mockRepo.On("Create", mock.AnythingOfType("*models.Post")).Return(errors.New("db error"))

	req := &models.CreatePostRequest{
		Title:   "Test Post",
		Content: "This is a test post",
	}

	// ACT
	post, err := postService.CreatePost(req, 1)

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, post)
	assert.Equal(t, "db error", err.Error())

	mockRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

// TestCreatePost_TitleVacio: validación previa falla si title vacío
func TestCreatePost_TitleVacio(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	req := &models.CreatePostRequest{
		Title:   "", // empty title
		Content: "Contenido",
	}

	// ACT
	post, err := postService.CreatePost(req, 1)

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, post)
	assert.Equal(t, "title is required", err.Error())
	// Must not call repo or userRepo
	mockRepo.AssertNotCalled(t, "Create")
	mockUserRepo.AssertNotCalled(t, "FindByID")
}

// TestCreatePost_ContentVacio: validación previa falla si content vacío
func TestCreatePost_ContentVacio(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	req := &models.CreatePostRequest{
		Title:   "Test Post",
		Content: "", // empty content
	}

	// ACT
	post, err := postService.CreatePost(req, 1)

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, post)
	assert.Equal(t, "content is required", err.Error())

	mockRepo.AssertNotCalled(t, "Create")
	mockUserRepo.AssertNotCalled(t, "FindByID")
}

// TestDeletePost_Success tests successful deletion by the author
func TestDeletePost_Success(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	existingPost := &models.Post{
		ID:       1,
		Title:    "Test Post",
		Content:  "Content",
		UserID:   1, // post author
		Username: "testuser",
	}

	// Configure mocks
	mockRepo.On("FindByID", 1).Return(existingPost, nil)
	mockRepo.On("Delete", 1).Return(nil)

	// ACT: user 1 deletes their own post
	err := postService.DeletePost(1, 1)

	// ASSERT
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

// TestDeletePost_PostNoExiste tests deleting a non-existent post
func TestDeletePost_PostNoExiste(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	// Post does not exist
	mockRepo.On("FindByID", 999).Return(nil, nil)

	// ACT
	err := postService.DeletePost(999, 1)

	// ASSERT
	assert.Error(t, err)
	assert.Equal(t, "post not found", err.Error())

	// Must not attempt deletion
	mockRepo.AssertNotCalled(t, "Delete")
}

// TestDeletePost_NoEsAutor tests that only the author can delete
func TestDeletePost_NoEsAutor(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	existingPost := &models.Post{
		ID:       1,
		Title:    "Test Post",
		Content:  "Content",
		UserID:   1, // post author
		Username: "testuser",
	}

	mockRepo.On("FindByID", 1).Return(existingPost, nil)

	// ACT: user 2 attempts to delete user 1's post
	err := postService.DeletePost(1, 2)

	// ASSERT
	assert.Error(t, err)
	assert.Equal(t, "you do not have permission to delete this post", err.Error())

	// Must not call Delete due to insufficient permissions
	mockRepo.AssertNotCalled(t, "Delete")
	mockRepo.AssertExpectations(t)
}

// TestDeleteComment_Success tests successful deletion by the author
func TestDeleteComment_Success(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	existingPost := &models.Post{
		ID:       1,
		Title:    "Test Post",
		UserID:   1,
		Username: "testuser",
	}

	existingUser := &models.User{
		ID:       1,
		Email:    "test@example.com",
		Username: "testuser",
	}

	// Configure mocks
	mockRepo.On("FindByID", 1).Return(existingPost, nil)
	mockUserRepo.On("FindByID", 1).Return(existingUser, nil)
	mockRepo.On("DeleteComment", 1, 10, 1).Return(nil)

	// ACT: user 1 deletes their own comment
	err := postService.DeleteComment(1, 10, 1)

	// ASSERT
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

// TestDeleteComment_PostNoExiste tests deleting a comment on a non-existent post
func TestDeleteComment_PostNoExiste(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	// Post does not exist
	mockRepo.On("FindByID", 999).Return(nil, nil)

	// ACT
	err := postService.DeleteComment(999, 10, 1)

	// ASSERT
	assert.Error(t, err)
	assert.Equal(t, "post not found", err.Error())

	// Must not attempt deletion
	mockRepo.AssertNotCalled(t, "DeleteComment")
}

// TestDeleteComment_UsuarioNoExiste tests deletion with non-existent user
func TestDeleteComment_UsuarioNoExiste(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	existingPost := &models.Post{
		ID:       1,
		Title:    "Test Post",
		UserID:   1,
		Username: "testuser",
	}

	mockRepo.On("FindByID", 1).Return(existingPost, nil)
	mockUserRepo.On("FindByID", 999).Return(nil, nil)

	// ACT
	err := postService.DeleteComment(1, 10, 999)

	// ASSERT
	assert.Error(t, err)
	assert.Equal(t, "user not found", err.Error())
	mockRepo.AssertNotCalled(t, "DeleteComment")
}

// TestDeleteComment_NoEsAutor tests that only the author can delete their comment
func TestDeleteComment_NoEsAutor(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	existingPost := &models.Post{
		ID:       1,
		Title:    "Test Post",
		UserID:   1,
		Username: "testuser",
	}

	existingUser := &models.User{
		ID:       2,
		Email:    "other@example.com",
		Username: "otheruser",
	}

	mockRepo.On("FindByID", 1).Return(existingPost, nil)
	mockUserRepo.On("FindByID", 2).Return(existingUser, nil)

	// User 2 attempts to delete user 1's comment
	mockRepo.On("DeleteComment", 1, 10, 2).Return(errors.New("you do not have permission to delete this comment or it does not exist"))

	// ACT
	err := postService.DeleteComment(1, 10, 2)

	// ASSERT
	assert.Error(t, err)
	assert.Equal(t, "you do not have permission to delete this comment or it does not exist", err.Error())
	mockRepo.AssertExpectations(t)
}

// TestGetAllPosts_Success tests retrieving all posts
func TestGetAllPosts_Success(t *testing.T) {
	// ARRANGE
	mockPostRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockPostRepo, mockUserRepo)

	mockPosts := []*models.Post{
		{ID: 1, Title: "Post 1", Content: "Content 1", UserID: 1},
		{ID: 2, Title: "Post 2", Content: "Content 2", UserID: 2},
	}
	mockPostRepo.On("FindAll").Return(mockPosts, nil)

	// ACT
	posts, err := postService.GetAllPosts()

	// ASSERT
	assert.NoError(t, err)
	assert.Len(t, posts, 2)
	mockPostRepo.AssertExpectations(t)
}

// TestGetPostByID_Success tests retrieving a post by ID
func TestGetPostByID_Success(t *testing.T) {
	// ARRANGE
	mockPostRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockPostRepo, mockUserRepo)

	mockPost := &models.Post{
		ID:      1,
		Title:   "Test Post",
		Content: "Test Content",
		UserID:  1,
	}
	mockPostRepo.On("FindByID", 1).Return(mockPost, nil)

	// ACT
	post, err := postService.GetPostByID(1)

	// ASSERT
	assert.NoError(t, err)
	assert.NotNil(t, post)
	assert.Equal(t, "Test Post", post.Title)
	mockPostRepo.AssertExpectations(t)
}

// TestCreateComment_Success tests creating a comment
func TestCreateComment_Success(t *testing.T) {
	// ARRANGE
	mockPostRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockPostRepo, mockUserRepo)

	mockPost := &models.Post{ID: 1, Title: "Post", UserID: 1}
	mockUser := &models.User{ID: 2, Username: "commenter"}

	mockPostRepo.On("FindByID", 1).Return(mockPost, nil)
	mockUserRepo.On("FindByID", 2).Return(mockUser, nil)
	mockPostRepo.On("CreateComment", mock.AnythingOfType("*models.Comment")).Return(nil)

	req := &models.CreateCommentRequest{
		Content: "Great post!",
	}

	// ACT
	comment, err := postService.CreateComment(1, req, 2)

	// ASSERT
	assert.NoError(t, err)
	assert.NotNil(t, comment)
	assert.Equal(t, "Great post!", comment.Content)
	assert.Equal(t, 1, comment.PostID)
	assert.Equal(t, 2, comment.UserID)
	mockPostRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

// TestGetCommentsByPostID_Success tests retrieving comments for a post
func TestGetCommentsByPostID_Success(t *testing.T) {
	// ARRANGE
	mockPostRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockPostRepo, mockUserRepo)

	mockPost := &models.Post{ID: 1, Title: "Post", UserID: 1}
	mockComments := []*models.Comment{
		{ID: 1, PostID: 1, UserID: 1, Content: "Comment 1"},
		{ID: 2, PostID: 1, UserID: 2, Content: "Comment 2"},
	}

	mockPostRepo.On("FindByID", 1).Return(mockPost, nil)
	mockPostRepo.On("FindCommentsByPostID", 1).Return(mockComments, nil)

	// ACT
	comments, err := postService.GetCommentsByPostID(1)

	// ASSERT
	assert.NoError(t, err)
	assert.Len(t, comments, 2)
	mockPostRepo.AssertExpectations(t)
}

// ========== Additional tests for GetAllPosts ==========

// TestGetAllPosts_Empty tests when there are no posts
func TestGetAllPosts_Empty(t *testing.T) {
	// ARRANGE
	mockPostRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockPostRepo, mockUserRepo)

	mockPostRepo.On("FindAll").Return(nil, nil)

	// ACT
	posts, err := postService.GetAllPosts()

	// ASSERT
	assert.NoError(t, err)
	assert.NotNil(t, posts)
	assert.Len(t, posts, 0)
	mockPostRepo.AssertExpectations(t)
}

// ========== Additional tests for GetPostByID ==========

// TestGetPostByID_InvalidID tests with invalid ID (negative or zero)
func TestGetPostByID_InvalidID(t *testing.T) {
	// ARRANGE
	mockPostRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockPostRepo, mockUserRepo)

	// ACT
	post, err := postService.GetPostByID(0)

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, post)
	assert.Equal(t, "invalid id", err.Error())
	mockPostRepo.AssertNotCalled(t, "FindByID")
}

// TestGetPostByID_NotFound tests when the post does not exist
func TestGetPostByID_NotFound(t *testing.T) {
	// ARRANGE
	mockPostRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockPostRepo, mockUserRepo)

	mockPostRepo.On("FindByID", 999).Return(nil, nil)

	// ACT
	post, err := postService.GetPostByID(999)

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, post)
	assert.Equal(t, "post not found", err.Error())
	mockPostRepo.AssertExpectations(t)
}

// ========== Additional tests for CreateComment ==========

// TestCreateComment_EmptyContent tests comment with empty content
func TestCreateComment_EmptyContent(t *testing.T) {
	// ARRANGE
	mockPostRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockPostRepo, mockUserRepo)

	req := &models.CreateCommentRequest{
		Content: "",
	}

	// ACT
	comment, err := postService.CreateComment(1, req, 1)

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, comment)
	assert.Equal(t, "comment content is required", err.Error())
	mockPostRepo.AssertNotCalled(t, "FindByID")
	mockPostRepo.AssertNotCalled(t, "CreateComment")
}

// TestCreateComment_PostNotFound tests commenting on non-existent post
func TestCreateComment_PostNotFound(t *testing.T) {
	// ARRANGE
	mockPostRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockPostRepo, mockUserRepo)

	mockPostRepo.On("FindByID", 999).Return(nil, nil)

	req := &models.CreateCommentRequest{
		Content: "Great post!",
	}

	// ACT
	comment, err := postService.CreateComment(999, req, 1)

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, comment)
	assert.Equal(t, "post not found", err.Error())
	mockPostRepo.AssertExpectations(t)
	mockPostRepo.AssertNotCalled(t, "CreateComment")
}

// TestCreateComment_UserNotFound tests commenting with non-existent user
func TestCreateComment_UserNotFound(t *testing.T) {
	// ARRANGE
	mockPostRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockPostRepo, mockUserRepo)

	mockPost := &models.Post{ID: 1, Title: "Post", UserID: 1}
	mockPostRepo.On("FindByID", 1).Return(mockPost, nil)
	mockUserRepo.On("FindByID", 999).Return(nil, nil)

	req := &models.CreateCommentRequest{
		Content: "Great post!",
	}

	// ACT
	comment, err := postService.CreateComment(1, req, 999)

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, comment)
	assert.Equal(t, "user not found", err.Error())
	mockPostRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockPostRepo.AssertNotCalled(t, "CreateComment")
}

// ========== Additional tests for GetCommentsByPostID ==========

// TestGetCommentsByPostID_PostNotFound tests retrieving comments from non-existent post
func TestGetCommentsByPostID_PostNotFound(t *testing.T) {
	// ARRANGE
	mockPostRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockPostRepo, mockUserRepo)

	mockPostRepo.On("FindByID", 999).Return(nil, nil)

	// ACT
	comments, err := postService.GetCommentsByPostID(999)

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, comments)
	assert.Equal(t, "post not found", err.Error())
	mockPostRepo.AssertExpectations(t)
	mockPostRepo.AssertNotCalled(t, "FindCommentsByPostID")
}

// TestGetCommentsByPostID_Empty tests when there are no comments
func TestGetCommentsByPostID_Empty(t *testing.T) {
	// ARRANGE
	mockPostRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockPostRepo, mockUserRepo)

	mockPost := &models.Post{ID: 1, Title: "Post", UserID: 1}
	mockPostRepo.On("FindByID", 1).Return(mockPost, nil)
	mockPostRepo.On("FindCommentsByPostID", 1).Return(nil, nil)

	// ACT
	comments, err := postService.GetCommentsByPostID(1)

	// ASSERT
	assert.NoError(t, err)
	assert.NotNil(t, comments)
	assert.Len(t, comments, 0)
	mockPostRepo.AssertExpectations(t)
}

// ========== Error propagation tests ==========

func TestCreatePost_ShouldReturnError_WhenTitleTooShort(t *testing.T) {
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	req := &models.CreatePostRequest{Title: "ab", Content: "Some content"}

	post, err := postService.CreatePost(req, 1)

	assert.Error(t, err)
	assert.Nil(t, post)
	assert.Equal(t, "title must be at least 3 characters", err.Error())
	mockRepo.AssertNotCalled(t, "Create")
	mockUserRepo.AssertNotCalled(t, "FindByID")
}

func TestCreatePost_ShouldPropagateError_WhenUserRepoFails(t *testing.T) {
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	mockUserRepo.On("FindByID", 1).Return(nil, errors.New("db error"))
	req := &models.CreatePostRequest{Title: "Valid Title", Content: "Some content"}

	post, err := postService.CreatePost(req, 1)

	assert.Error(t, err)
	assert.Nil(t, post)
	assert.Equal(t, "db error", err.Error())
	mockUserRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "Create")
}

func TestGetAllPosts_ShouldPropagateError_WhenRepoFails(t *testing.T) {
	mockPostRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockPostRepo, mockUserRepo)

	mockPostRepo.On("FindAll").Return(nil, errors.New("db error"))

	posts, err := postService.GetAllPosts()

	assert.Error(t, err)
	assert.Nil(t, posts)
	assert.Equal(t, "db error", err.Error())
	mockPostRepo.AssertExpectations(t)
}

func TestGetPostByID_ShouldPropagateError_WhenRepoFails(t *testing.T) {
	mockPostRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockPostRepo, mockUserRepo)

	mockPostRepo.On("FindByID", 1).Return(nil, errors.New("db error"))

	post, err := postService.GetPostByID(1)

	assert.Error(t, err)
	assert.Nil(t, post)
	assert.Equal(t, "db error", err.Error())
	mockPostRepo.AssertExpectations(t)
}

func TestDeletePost_ShouldPropagateError_WhenRepoFails(t *testing.T) {
	mockPostRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockPostRepo, mockUserRepo)

	mockPostRepo.On("FindByID", 1).Return(nil, errors.New("db error"))

	err := postService.DeletePost(1, 1)

	assert.Error(t, err)
	assert.Equal(t, "db error", err.Error())
	mockPostRepo.AssertExpectations(t)
	mockPostRepo.AssertNotCalled(t, "Delete")
}

func TestCreateComment_ShouldPropagateError_WhenPostRepoFails(t *testing.T) {
	mockPostRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockPostRepo, mockUserRepo)

	mockPostRepo.On("FindByID", 1).Return(nil, errors.New("db error"))
	req := &models.CreateCommentRequest{Content: "A comment"}

	comment, err := postService.CreateComment(1, req, 1)

	assert.Error(t, err)
	assert.Nil(t, comment)
	assert.Equal(t, "db error", err.Error())
	mockPostRepo.AssertExpectations(t)
	mockUserRepo.AssertNotCalled(t, "FindByID")
	mockPostRepo.AssertNotCalled(t, "CreateComment")
}

func TestCreateComment_ShouldPropagateError_WhenUserRepoFails(t *testing.T) {
	mockPostRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockPostRepo, mockUserRepo)

	mockPost := &models.Post{ID: 1, Title: "Test Post", UserID: 1}
	mockPostRepo.On("FindByID", 1).Return(mockPost, nil)
	mockUserRepo.On("FindByID", 1).Return(nil, errors.New("db error"))
	req := &models.CreateCommentRequest{Content: "A comment"}

	comment, err := postService.CreateComment(1, req, 1)

	assert.Error(t, err)
	assert.Nil(t, comment)
	assert.Equal(t, "db error", err.Error())
	mockPostRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockPostRepo.AssertNotCalled(t, "CreateComment")
}

func TestCreateComment_ShouldPropagateError_WhenCreateCommentFails(t *testing.T) {
	mockPostRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockPostRepo, mockUserRepo)

	mockPost := &models.Post{ID: 1, Title: "Test Post", UserID: 1}
	mockUser := &models.User{ID: 1, Username: "testuser"}
	mockPostRepo.On("FindByID", 1).Return(mockPost, nil)
	mockUserRepo.On("FindByID", 1).Return(mockUser, nil)
	mockPostRepo.On("CreateComment", mock.AnythingOfType("*models.Comment")).Return(errors.New("db error"))
	req := &models.CreateCommentRequest{Content: "A comment"}

	comment, err := postService.CreateComment(1, req, 1)

	assert.Error(t, err)
	assert.Nil(t, comment)
	assert.Equal(t, "db error", err.Error())
	mockPostRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestGetCommentsByPostID_ShouldPropagateError_WhenPostRepoFails(t *testing.T) {
	mockPostRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockPostRepo, mockUserRepo)

	mockPostRepo.On("FindByID", 1).Return(nil, errors.New("db error"))

	comments, err := postService.GetCommentsByPostID(1)

	assert.Error(t, err)
	assert.Nil(t, comments)
	assert.Equal(t, "db error", err.Error())
	mockPostRepo.AssertExpectations(t)
	mockPostRepo.AssertNotCalled(t, "FindCommentsByPostID")
}

func TestGetCommentsByPostID_ShouldPropagateError_WhenFindCommentsFails(t *testing.T) {
	mockPostRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockPostRepo, mockUserRepo)

	mockPost := &models.Post{ID: 1, Title: "Test Post", UserID: 1}
	mockPostRepo.On("FindByID", 1).Return(mockPost, nil)
	mockPostRepo.On("FindCommentsByPostID", 1).Return(nil, errors.New("db error"))

	comments, err := postService.GetCommentsByPostID(1)

	assert.Error(t, err)
	assert.Nil(t, comments)
	assert.Equal(t, "db error", err.Error())
	mockPostRepo.AssertExpectations(t)
}

func TestDeleteComment_ShouldPropagateError_WhenPostRepoFails(t *testing.T) {
	mockPostRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockPostRepo, mockUserRepo)

	mockPostRepo.On("FindByID", 1).Return(nil, errors.New("db error"))

	err := postService.DeleteComment(1, 10, 1)

	assert.Error(t, err)
	assert.Equal(t, "db error", err.Error())
	mockPostRepo.AssertExpectations(t)
	mockUserRepo.AssertNotCalled(t, "FindByID")
	mockPostRepo.AssertNotCalled(t, "DeleteComment")
}

func TestDeleteComment_ShouldPropagateError_WhenUserRepoFails(t *testing.T) {
	mockPostRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockPostRepo, mockUserRepo)

	mockPost := &models.Post{ID: 1, Title: "Test Post", UserID: 1}
	mockPostRepo.On("FindByID", 1).Return(mockPost, nil)
	mockUserRepo.On("FindByID", 1).Return(nil, errors.New("db error"))

	err := postService.DeleteComment(1, 10, 1)

	assert.Error(t, err)
	assert.Equal(t, "db error", err.Error())
	mockPostRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockPostRepo.AssertNotCalled(t, "DeleteComment")
}
