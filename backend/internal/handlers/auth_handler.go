package handlers

import (
	"encoding/json"
	"net/http"

	"tp06-testing/internal/models"
	"tp06-testing/internal/services"
)

// AuthHandler handles HTTP authentication requests
type AuthHandler struct {
	authService *services.AuthService
}

// NewAuthHandler creates a new instance
func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register handles POST /api/auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	// Decode JSON body
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	// Call the service
	user, err := h.authService.Register(&req)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Respond with the created user
	respondWithJSON(w, http.StatusCreated, user)
}

// Login handles POST /api/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// Decode JSON body
	var creds models.Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	// Call the service
	user, err := h.authService.Login(&creds)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Respond with the authenticated user
	respondWithJSON(w, http.StatusOK, user)
}

// Helper functions for JSON responses

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}
