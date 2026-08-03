package models

import "time"

// Post represents a forum post
type Post struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	UserID    int       `json:"user_id"`
	Username  string    `json:"username"` // To display who authored it
	CreatedAt time.Time `json:"created_at"`
}

// CreatePostRequest is used to create a post
type CreatePostRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

// EditPostRequest is used to edit an existing post
type EditPostRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

// Comment represents a comment on a post
type Comment struct {
	ID        int       `json:"id"`
	PostID    int       `json:"post_id"`
	UserID    int       `json:"user_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateCommentRequest is used to create a comment
type CreateCommentRequest struct {
	Content string `json:"content"`
}

// EditCommentRequest is used to edit an existing comment
type EditCommentRequest struct {
	Content string `json:"content"`
}
