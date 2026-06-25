package models

import "time"

// Post representa una publicación
type Post struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	UserID    int       `json:"user_id"`
	Username  string    `json:"username"` // Para mostrar quién publicó
	CreatedAt time.Time `json:"created_at"`
}

// CreatePostRequest se usa para crear un post
type CreatePostRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

// Comment representa un comentario en un post
type Comment struct {
	ID        int       `json:"id"`
	PostID    int       `json:"post_id"`
	UserID    int       `json:"user_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateCommentRequest se usa para crear un comentario
type CreateCommentRequest struct {
	Content string `json:"content"`
}