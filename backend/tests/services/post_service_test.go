package services

import (
	"errors"
	"testing"

	"forum-app-qa-pipeline/internal/models"
	"forum-app-qa-pipeline/internal/services"
	"forum-app-qa-pipeline/tests/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestCreatePost_Success verifies a post is created successfully
func TestCreatePost_Success(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	existingUser := &models.User{
		ID:       1,
		Email:    "test@example.com",
		Username: "testuser",
	}
	mockUserRepo.On("FindByID", 1).Return(existingUser, nil)

	// Configure the mock: Create should execute successfully
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

	// Verify the mock's methods were called
	mockRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

// TestCreatePost_UserNotFound verifies creation fails when the given userId does not exist
func TestCreatePost_UserNotFound(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	// Configure the mock: FindByID returns nil (user does not exist)
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

// TestCreatePost_RepoError verifies the error is propagated when the repository fails to create the post
func TestCreatePost_RepoError(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	// The user exists
	existingUser := &models.User{ID: 1, Email: "u@u.com", Username: "u"}
	mockUserRepo.On("FindByID", 1).Return(existingUser, nil)

	// The repository's Create call fails
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

// TestCreatePost_EmptyTitle verifies the pre-check fails when the title is empty
func TestCreatePost_EmptyTitle(t *testing.T) {
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
	// Should NOT call the repo or the userRepo
	mockRepo.AssertNotCalled(t, "Create")
	mockUserRepo.AssertNotCalled(t, "FindByID")
}

// TestCreatePost_EmptyContent verifies the pre-check fails when the content is empty
func TestCreatePost_EmptyContent(t *testing.T) {
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

// TestGetAllPosts_Success verifies all posts are returned as-is
func TestGetAllPosts_Success(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	existingPosts := []*models.Post{
		{ID: 1, Title: "First Post", Content: "Content 1", UserID: 1, Username: "user1"},
		{ID: 2, Title: "Second Post", Content: "Content 2", UserID: 2, Username: "user2"},
	}
	mockRepo.On("FindAll").Return(existingPosts, nil)

	// ACT
	posts, err := postService.GetAllPosts()

	// ASSERT
	assert.NoError(t, err)
	assert.Equal(t, existingPosts, posts)
	mockRepo.AssertExpectations(t)
}

// TestGetAllPosts_Empty verifies an empty list is returned, not nil, when there are no posts
func TestGetAllPosts_Empty(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	mockRepo.On("FindAll").Return(nil, nil)

	// ACT
	posts, err := postService.GetAllPosts()

	// ASSERT
	assert.NoError(t, err)
	assert.NotNil(t, posts)
	assert.Empty(t, posts)
}

// TestGetAllPosts_RepoError verifies the error is propagated when the repository fails
func TestGetAllPosts_RepoError(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	mockRepo.On("FindAll").Return(nil, errors.New("db error"))

	// ACT
	posts, err := postService.GetAllPosts()

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, posts)
	assert.Equal(t, "db error", err.Error())
}

// TestGetPostByID_Success verifies a post is returned when the id is valid and exists
func TestGetPostByID_Success(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	existingPost := &models.Post{ID: 1, Title: "Test Post", Content: "Content", UserID: 1, Username: "testuser"}
	mockRepo.On("FindByID", 1).Return(existingPost, nil)

	// ACT
	post, err := postService.GetPostByID(1)

	// ASSERT
	assert.NoError(t, err)
	assert.Equal(t, existingPost, post)
	mockRepo.AssertExpectations(t)
}

// TestGetPostByID_InvalidID verifies the pre-check fails for a non-positive id
func TestGetPostByID_InvalidID(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	// ACT
	post, err := postService.GetPostByID(0)

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, post)
	assert.Equal(t, "invalid id", err.Error())
	mockRepo.AssertNotCalled(t, "FindByID")
}

// TestGetPostByID_NotFound verifies the lookup fails when the post does not exist
func TestGetPostByID_NotFound(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	mockRepo.On("FindByID", 999).Return(nil, nil)

	// ACT
	post, err := postService.GetPostByID(999)

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, post)
	assert.Equal(t, "post not found", err.Error())
}

// TestGetPostByID_RepoError verifies the error is propagated when the repository fails
func TestGetPostByID_RepoError(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	mockRepo.On("FindByID", 1).Return(nil, errors.New("db error"))

	// ACT
	post, err := postService.GetPostByID(1)

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, post)
	assert.Equal(t, "db error", err.Error())
}

// TestEditPost_Success verifies the author can edit their own post
func TestEditPost_Success(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	existingPost := &models.Post{
		ID:       1,
		Title:    "Original Title",
		Content:  "Original content",
		UserID:   1, // The author is user 1
		Username: "testuser",
	}

	mockRepo.On("FindByID", 1).Return(existingPost, nil)
	mockRepo.On("Update", mock.AnythingOfType("*models.Post")).Return(nil)

	req := &models.EditPostRequest{
		Title:   "Updated Title",
		Content: "Updated content",
	}

	// ACT
	post, err := postService.EditPost(1, req, 1)

	// ASSERT
	assert.NoError(t, err)
	assert.NotNil(t, post)
	assert.Equal(t, "Updated Title", post.Title)
	assert.Equal(t, "Updated content", post.Content)

	mockRepo.AssertExpectations(t)
}

// TestEditPost_EmptyFields verifies the pre-check fails when title or content is empty
func TestEditPost_EmptyFields(t *testing.T) {
	tests := []struct {
		name    string
		title   string
		content string
		wantErr string
	}{
		{"EmptyTitle", "", "Updated content", "title is required"},
		{"EmptyContent", "Updated Title", "", "content is required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ARRANGE
			mockRepo := new(mocks.MockPostRepository)
			mockUserRepo := new(mocks.MockUserRepository)
			postService := services.NewPostService(mockRepo, mockUserRepo)

			req := &models.EditPostRequest{
				Title:   tt.title,
				Content: tt.content,
			}

			// ACT
			post, err := postService.EditPost(1, req, 1)

			// ASSERT
			assert.Error(t, err)
			assert.Nil(t, post)
			assert.Equal(t, tt.wantErr, err.Error())
			// Should NOT even look up the post
			mockRepo.AssertNotCalled(t, "FindByID")
			mockRepo.AssertNotCalled(t, "Update")
		})
	}
}

// TestEditPost_PostNotFound verifies the edit fails when the post does not exist
func TestEditPost_PostNotFound(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	// The post does not exist
	mockRepo.On("FindByID", 999).Return(nil, nil)

	req := &models.EditPostRequest{
		Title:   "Updated Title",
		Content: "Updated content",
	}

	// ACT
	post, err := postService.EditPost(999, req, 1)

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, post)
	assert.Equal(t, "post not found", err.Error())
	mockRepo.AssertNotCalled(t, "Update")
}

// TestEditPost_NotTheAuthor verifies only the post's author can edit it
func TestEditPost_NotTheAuthor(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	existingPost := &models.Post{
		ID:       1,
		Title:    "Original Title",
		Content:  "Original content",
		UserID:   1, // The author is user 1
		Username: "testuser",
	}

	mockRepo.On("FindByID", 1).Return(existingPost, nil)

	req := &models.EditPostRequest{
		Title:   "Updated Title",
		Content: "Updated content",
	}

	// User 2 attempts to edit user 1's post
	// ACT
	post, err := postService.EditPost(1, req, 2)

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, post)
	assert.Equal(t, "you do not have permission to edit this post", err.Error())
	mockRepo.AssertNotCalled(t, "Update")
	mockRepo.AssertExpectations(t)
}

// TestDeletePost_Success verifies the author can delete their own post
func TestDeletePost_Success(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	existingPost := &models.Post{
		ID:       1,
		Title:    "Test Post",
		Content:  "Content",
		UserID:   1, // The author is user 1
		Username: "testuser",
	}

	// Configure mocks
	mockRepo.On("FindByID", 1).Return(existingPost, nil)
	mockRepo.On("Delete", 1).Return(nil)

	// User 1 deletes their own post
	// ACT
	err := postService.DeletePost(1, 1)

	// ASSERT
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

// TestDeletePost_PostNotFound verifies deletion fails when the post does not exist
func TestDeletePost_PostNotFound(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	// The post does not exist
	mockRepo.On("FindByID", 999).Return(nil, nil)

	// ACT
	err := postService.DeletePost(999, 1)

	// ASSERT
	assert.Error(t, err)
	assert.Equal(t, "post not found", err.Error())

	// Should NOT attempt to delete
	mockRepo.AssertNotCalled(t, "Delete")
}

// TestDeletePost_NotTheAuthor verifies only the post's author can delete it
func TestDeletePost_NotTheAuthor(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	existingPost := &models.Post{
		ID:       1,
		Title:    "Test Post",
		Content:  "Content",
		UserID:   1, // The author is user 1
		Username: "testuser",
	}

	mockRepo.On("FindByID", 1).Return(existingPost, nil)

	// User 2 attempts to delete user 1's post
	// ACT
	err := postService.DeletePost(1, 2)

	// ASSERT
	assert.Error(t, err)
	assert.Equal(t, "you do not have permission to delete this post", err.Error())

	// Should NOT call Delete because they don't have permission
	mockRepo.AssertNotCalled(t, "Delete")
	mockRepo.AssertExpectations(t)
}

// TestCreateComment_Success verifies a comment is created successfully
func TestCreateComment_Success(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	existingPost := &models.Post{ID: 1, Title: "Test Post", UserID: 1, Username: "testuser"}
	existingUser := &models.User{ID: 2, Email: "u2@u.com", Username: "user2"}

	mockRepo.On("FindByID", 1).Return(existingPost, nil)
	mockUserRepo.On("FindByID", 2).Return(existingUser, nil)
	mockRepo.On("CreateComment", mock.AnythingOfType("*models.Comment")).Return(nil)

	req := &models.CreateCommentRequest{Content: "A comment"}

	// ACT
	comment, err := postService.CreateComment(1, req, 2)

	// ASSERT
	assert.NoError(t, err)
	assert.NotNil(t, comment)
	assert.Equal(t, "A comment", comment.Content)
	assert.Equal(t, "user2", comment.Username)
	mockRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

// TestCreateComment_EmptyContent verifies the pre-check fails when the content is empty
func TestCreateComment_EmptyContent(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	req := &models.CreateCommentRequest{Content: ""}

	// ACT
	comment, err := postService.CreateComment(1, req, 1)

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, comment)
	assert.Equal(t, "comment content is required", err.Error())
	mockRepo.AssertNotCalled(t, "FindByID")
	mockRepo.AssertNotCalled(t, "CreateComment")
}

// TestCreateComment_PostNotFound verifies creation fails when the parent post does not exist
func TestCreateComment_PostNotFound(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	mockRepo.On("FindByID", 999).Return(nil, nil)

	req := &models.CreateCommentRequest{Content: "A comment"}

	// ACT
	comment, err := postService.CreateComment(999, req, 1)

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, comment)
	assert.Equal(t, "post not found", err.Error())
	mockUserRepo.AssertNotCalled(t, "FindByID")
	mockRepo.AssertNotCalled(t, "CreateComment")
}

// TestCreateComment_UserNotFound verifies creation fails when the requesting user does not exist
func TestCreateComment_UserNotFound(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	existingPost := &models.Post{ID: 1, Title: "Test Post", UserID: 1, Username: "testuser"}
	mockRepo.On("FindByID", 1).Return(existingPost, nil)
	mockUserRepo.On("FindByID", 999).Return(nil, nil)

	req := &models.CreateCommentRequest{Content: "A comment"}

	// ACT
	comment, err := postService.CreateComment(1, req, 999)

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, comment)
	assert.Equal(t, "user not found", err.Error())
	mockRepo.AssertNotCalled(t, "CreateComment")
}

// TestCreateComment_RepoError verifies the error is propagated when the repository fails to create the comment
func TestCreateComment_RepoError(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	existingPost := &models.Post{ID: 1, Title: "Test Post", UserID: 1, Username: "testuser"}
	existingUser := &models.User{ID: 1, Email: "u@u.com", Username: "testuser"}

	mockRepo.On("FindByID", 1).Return(existingPost, nil)
	mockUserRepo.On("FindByID", 1).Return(existingUser, nil)
	mockRepo.On("CreateComment", mock.AnythingOfType("*models.Comment")).Return(errors.New("db error"))

	req := &models.CreateCommentRequest{Content: "A comment"}

	// ACT
	comment, err := postService.CreateComment(1, req, 1)

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, comment)
	assert.Equal(t, "db error", err.Error())
}

// TestGetCommentsByPostID_Success verifies all comments for a post are returned
func TestGetCommentsByPostID_Success(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	existingPost := &models.Post{ID: 1, Title: "Test Post", UserID: 1, Username: "testuser"}
	existingComments := []*models.Comment{
		{ID: 1, PostID: 1, UserID: 1, Username: "testuser", Content: "First comment"},
		{ID: 2, PostID: 1, UserID: 2, Username: "otheruser", Content: "Second comment"},
	}

	mockRepo.On("FindByID", 1).Return(existingPost, nil)
	mockRepo.On("FindCommentsByPostID", 1).Return(existingComments, nil)

	// ACT
	comments, err := postService.GetCommentsByPostID(1)

	// ASSERT
	assert.NoError(t, err)
	assert.Equal(t, existingComments, comments)
	mockRepo.AssertExpectations(t)
}

// TestGetCommentsByPostID_PostNotFound verifies the lookup fails when the parent post does not exist
func TestGetCommentsByPostID_PostNotFound(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	mockRepo.On("FindByID", 999).Return(nil, nil)

	// ACT
	comments, err := postService.GetCommentsByPostID(999)

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, comments)
	assert.Equal(t, "post not found", err.Error())
	mockRepo.AssertNotCalled(t, "FindCommentsByPostID")
}

// TestGetCommentsByPostID_Empty verifies an empty list is returned, not nil, when there are no comments
func TestGetCommentsByPostID_Empty(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	existingPost := &models.Post{ID: 1, Title: "Test Post", UserID: 1, Username: "testuser"}
	mockRepo.On("FindByID", 1).Return(existingPost, nil)
	mockRepo.On("FindCommentsByPostID", 1).Return(nil, nil)

	// ACT
	comments, err := postService.GetCommentsByPostID(1)

	// ASSERT
	assert.NoError(t, err)
	assert.NotNil(t, comments)
	assert.Empty(t, comments)
}

// TestGetCommentsByPostID_RepoError verifies the error is propagated when the repository fails
func TestGetCommentsByPostID_RepoError(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	existingPost := &models.Post{ID: 1, Title: "Test Post", UserID: 1, Username: "testuser"}
	mockRepo.On("FindByID", 1).Return(existingPost, nil)
	mockRepo.On("FindCommentsByPostID", 1).Return(nil, errors.New("db error"))

	// ACT
	comments, err := postService.GetCommentsByPostID(1)

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, comments)
	assert.Equal(t, "db error", err.Error())
}

// newEditCommentFixture returns a post and a comment on it, both authored by
// user 1 -- shared by EditComment tests that need an existing, editable
// comment as their starting state.
func newEditCommentFixture() (*models.Post, *models.Comment) {
	post := &models.Post{ID: 1, Title: "Test Post", UserID: 1, Username: "testuser"}
	comment := &models.Comment{
		ID:       10,
		PostID:   1,
		UserID:   1, // The author is user 1
		Username: "testuser",
		Content:  "Original content",
	}
	return post, comment
}

// TestEditComment_Success verifies the author can edit their own comment
func TestEditComment_Success(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	existingPost, existingComment := newEditCommentFixture()

	mockRepo.On("FindByID", 1).Return(existingPost, nil)
	mockRepo.On("FindCommentByID", 10).Return(existingComment, nil)
	mockRepo.On("UpdateComment", mock.AnythingOfType("*models.Comment")).Return(nil)

	req := &models.EditCommentRequest{Content: "Updated content"}

	// User 1 edits their own comment
	// ACT
	comment, err := postService.EditComment(1, 10, req, 1)

	// ASSERT
	assert.NoError(t, err)
	assert.NotNil(t, comment)
	assert.Equal(t, "Updated content", comment.Content)
	mockRepo.AssertExpectations(t)
}

// TestEditComment_EmptyContent verifies the pre-check fails when the content is empty
func TestEditComment_EmptyContent(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	req := &models.EditCommentRequest{Content: ""} // empty content

	// ACT
	comment, err := postService.EditComment(1, 10, req, 1)

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, comment)
	assert.Equal(t, "comment content is required", err.Error())
	mockRepo.AssertNotCalled(t, "FindByID")
	mockRepo.AssertNotCalled(t, "UpdateComment")
}

// TestEditComment_PostNotFound verifies the edit fails when the parent post does not exist
func TestEditComment_PostNotFound(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	// The post does not exist
	mockRepo.On("FindByID", 999).Return(nil, nil)

	req := &models.EditCommentRequest{Content: "Updated content"}

	// ACT
	comment, err := postService.EditComment(999, 10, req, 1)

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, comment)
	assert.Equal(t, "post not found", err.Error())
	mockRepo.AssertNotCalled(t, "FindCommentByID")
	mockRepo.AssertNotCalled(t, "UpdateComment")
}

// TestEditComment_CommentNotFound verifies the edit fails when the comment does not exist
func TestEditComment_CommentNotFound(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	existingPost := &models.Post{ID: 1, Title: "Test Post", UserID: 1, Username: "testuser"}

	mockRepo.On("FindByID", 1).Return(existingPost, nil)
	mockRepo.On("FindCommentByID", 999).Return(nil, nil)

	req := &models.EditCommentRequest{Content: "Updated content"}

	// ACT
	comment, err := postService.EditComment(1, 999, req, 1)

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, comment)
	assert.Equal(t, "comment not found", err.Error())
	mockRepo.AssertNotCalled(t, "UpdateComment")
}

// TestEditComment_PostMismatch verifies the edit fails when the comment belongs to a different post
func TestEditComment_PostMismatch(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	existingPost := &models.Post{ID: 1, Title: "Test Post", UserID: 1, Username: "testuser"}
	// The comment exists, but belongs to a different post (post 2)
	existingComment := &models.Comment{ID: 10, PostID: 2, UserID: 1, Username: "testuser", Content: "Original content"}

	mockRepo.On("FindByID", 1).Return(existingPost, nil)
	mockRepo.On("FindCommentByID", 10).Return(existingComment, nil)

	req := &models.EditCommentRequest{Content: "Updated content"}

	// ACT
	comment, err := postService.EditComment(1, 10, req, 1)

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, comment)
	assert.Equal(t, "comment not found", err.Error())
	mockRepo.AssertNotCalled(t, "UpdateComment")
}

// TestEditComment_NotTheAuthor verifies only the comment's author can edit it
func TestEditComment_NotTheAuthor(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	existingPost, existingComment := newEditCommentFixture()

	mockRepo.On("FindByID", 1).Return(existingPost, nil)
	mockRepo.On("FindCommentByID", 10).Return(existingComment, nil)

	req := &models.EditCommentRequest{Content: "Updated content"}

	// User 2 attempts to edit user 1's comment
	// ACT
	comment, err := postService.EditComment(1, 10, req, 2)

	// ASSERT
	assert.Error(t, err)
	assert.Nil(t, comment)
	assert.Equal(t, "you do not have permission to edit this comment", err.Error())
	mockRepo.AssertNotCalled(t, "UpdateComment")
	mockRepo.AssertExpectations(t)
}

// TestDeleteComment_Success verifies the author can delete their own comment
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

	// User 1 deletes their own comment
	// ACT
	err := postService.DeleteComment(1, 10, 1)

	// ASSERT
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

// TestDeleteComment_PostNotFound verifies comment deletion fails when the parent post does not exist
func TestDeleteComment_PostNotFound(t *testing.T) {
	// ARRANGE
	mockRepo := new(mocks.MockPostRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	postService := services.NewPostService(mockRepo, mockUserRepo)

	// The post does not exist
	mockRepo.On("FindByID", 999).Return(nil, nil)

	// ACT
	err := postService.DeleteComment(999, 10, 1)

	// ASSERT
	assert.Error(t, err)
	assert.Equal(t, "post not found", err.Error())

	// Should NOT attempt to delete
	mockRepo.AssertNotCalled(t, "DeleteComment")
}

// TestDeleteComment_UserNotFound verifies comment deletion fails when the requesting user does not exist
func TestDeleteComment_UserNotFound(t *testing.T) {
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

// TestDeleteComment_NotTheAuthor verifies only the comment's author can delete it
func TestDeleteComment_NotTheAuthor(t *testing.T) {
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
