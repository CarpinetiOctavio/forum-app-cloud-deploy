package repository

import (
	"database/sql"
	"errors"

	"forum-app-cloud-deploy/internal/models"
)

// PostRepository defines operations on posts
type PostRepository interface {
	Create(post *models.Post) error
	FindAll() ([]*models.Post, error)
	FindByID(id int) (*models.Post, error)
	Delete(id int) error
	CreateComment(comment *models.Comment) error
	FindCommentsByPostID(postID int) ([]*models.Comment, error)
	DeleteComment(postID int, commentID int, userID int) error
}

// SQLitePostRepository implements PostRepository using SQLite
type SQLitePostRepository struct {
	db *sql.DB
}

// NewSQLitePostRepository creates a new instance
func NewSQLitePostRepository(db *sql.DB) *SQLitePostRepository {
	return &SQLitePostRepository{db: db}
}

// Create inserts a new post
func (r *SQLitePostRepository) Create(post *models.Post) error {
	query := `
		INSERT INTO posts (title, content, user_id, created_at)
		VALUES (?, ?, ?, datetime('now'))
	`
	result, err := r.db.Exec(query, post.Title, post.Content, post.UserID)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	post.ID = int(id)
	return nil
}

// FindAll retrieves all posts with author information
func (r *SQLitePostRepository) FindAll() ([]*models.Post, error) {
	query := `
		SELECT p.id, p.title, p.content, p.user_id, u.username, p.created_at
		FROM posts p
		JOIN users u ON p.user_id = u.id
		ORDER BY p.created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		post := &models.Post{}
		err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.UserID,
			&post.Username,
			&post.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	return posts, nil
}

// FindByID finds a post by ID
func (r *SQLitePostRepository) FindByID(id int) (*models.Post, error) {
	query := `
		SELECT p.id, p.title, p.content, p.user_id, u.username, p.created_at
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.id = ?
	`

	post := &models.Post{}
	err := r.db.QueryRow(query, id).Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.UserID,
		&post.Username,
		&post.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return post, nil
}

// Delete removes a post by ID
func (r *SQLitePostRepository) Delete(id int) error {
	query := `DELETE FROM posts WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

// CreateComment inserts a new comment
func (r *SQLitePostRepository) CreateComment(comment *models.Comment) error {
	query := `
		INSERT INTO comments (post_id, user_id, content, created_at)
		VALUES (?, ?, ?, datetime('now'))
	`
	result, err := r.db.Exec(query, comment.PostID, comment.UserID, comment.Content)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	comment.ID = int(id)
	return nil
}

// FindCommentsByPostID retrieves all comments for a post
func (r *SQLitePostRepository) FindCommentsByPostID(postID int) ([]*models.Comment, error) {
	query := `
		SELECT c.id, c.post_id, c.user_id, u.username, c.content, c.created_at
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.post_id = ?
		ORDER BY c.created_at ASC
	`

	rows, err := r.db.Query(query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*models.Comment
	for rows.Next() {
		comment := &models.Comment{}
		err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.UserID,
			&comment.Username,
			&comment.Content,
			&comment.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

func (r *SQLitePostRepository) DeleteComment(postID int, commentID int, userID int) error {
	query := `
		DELETE FROM comments
		WHERE id = ? AND post_id = ? AND user_id = ?
	`
	result, err := r.db.Exec(query, commentID, postID, userID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("you do not have permission to delete this comment or it does not exist")
	}
	return nil
}
