package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"tp06-testing/internal/models"
	"tp06-testing/internal/services"

	"github.com/gorilla/mux"
)

// Constants for headers and error messages
const (
	HeaderUserID            = "X-User-ID"
	ErrUserNotAuthenticated = "user not authenticated"
	ErrInvalidUserID        = "invalid user ID"
	ErrInvalidID            = "invalid ID"
	ErrInvalidJSON          = "invalid JSON"
)

// PostHandler handles HTTP post requests
type PostHandler struct {
	postService *services.PostService
}

// NewPostHandler creates a new instance
func NewPostHandler(postService *services.PostService) *PostHandler {
	return &PostHandler{
		postService: postService,
	}
}

// CreatePost handles POST /api/posts
func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	var req models.CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrInvalidJSON)
		return
	}

	userIDStr := r.Header.Get(HeaderUserID)
	if userIDStr == "" {
		respondWithError(w, http.StatusUnauthorized, ErrUserNotAuthenticated)
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, ErrInvalidUserID)
		return
	}

	post, err := h.postService.CreatePost(&req, userID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, post)
}

// GetAllPosts handles GET /api/posts
func (h *PostHandler) GetAllPosts(w http.ResponseWriter, r *http.Request) {
	posts, err := h.postService.GetAllPosts()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, posts)
}

// GetPostByID handles GET /api/posts/{id}
func (h *PostHandler) GetPostByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, ErrInvalidID)
		return
	}

	post, err := h.postService.GetPostByID(id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, post)
}

// DeletePost handles DELETE /api/posts/{id}
func (h *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, ErrInvalidID)
		return
	}

	userIDStr := r.Header.Get(HeaderUserID)
	if userIDStr == "" {
		respondWithError(w, http.StatusUnauthorized, ErrUserNotAuthenticated)
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, ErrInvalidUserID)
		return
	}

	err = h.postService.DeletePost(id, userID)
	if err != nil {
		respondWithError(w, http.StatusForbidden, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "post deleted"})
}

// CreateComment handles POST /api/posts/{id}/comments
func (h *PostHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, ErrInvalidID)
		return
	}

	var req models.CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrInvalidJSON)
		return
	}

	userIDStr := r.Header.Get(HeaderUserID)
	if userIDStr == "" {
		respondWithError(w, http.StatusUnauthorized, ErrUserNotAuthenticated)
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, ErrInvalidUserID)
		return
	}

	comment, err := h.postService.CreateComment(postID, &req, userID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, comment)
}

// GetComments handles GET /api/posts/{id}/comments
func (h *PostHandler) GetComments(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, ErrInvalidID)
		return
	}

	comments, err := h.postService.GetCommentsByPostID(postID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, comments)
}

// DeleteComment handles DELETE /api/posts/{postId}/comments/{commentId}
func (h *PostHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID, err := strconv.Atoi(vars["postId"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid post ID")
		return
	}
	commentID, err := strconv.Atoi(vars["commentId"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid comment ID")
		return
	}

	userIDStr := r.Header.Get(HeaderUserID)
	if userIDStr == "" {
		respondWithError(w, http.StatusUnauthorized, ErrUserNotAuthenticated)
		return
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, ErrInvalidUserID)
		return
	}

	err = h.postService.DeleteComment(postID, commentID, userID)
	if err != nil {
		respondWithError(w, http.StatusForbidden, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "comment deleted"})
}
