package repository

import (
	"database/sql"

	"forum-app-cloud-deploy/internal/models"
)

// UserRepository defines operations on users
// INTERFACE: allows easy mock creation for testing
type UserRepository interface {
	Create(user *models.User) error
	FindByEmail(email string) (*models.User, error)
	FindByID(id int) (*models.User, error)
}

// SQLiteUserRepository implements UserRepository using SQLite
type SQLiteUserRepository struct {
	db *sql.DB
}

// NewSQLiteUserRepository creates a new instance
func NewSQLiteUserRepository(db *sql.DB) *SQLiteUserRepository {
	return &SQLiteUserRepository{db: db}
}

// Create inserts a new user into the database
func (r *SQLiteUserRepository) Create(user *models.User) error {
	query := `
		INSERT INTO users (email, password, username, created_at)
		VALUES (?, ?, ?, datetime('now'))
	`
	result, err := r.db.Exec(query, user.Email, user.Password, user.Username)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	user.ID = int(id)
	return nil
}

// FindByEmail finds a user by email
func (r *SQLiteUserRepository) FindByEmail(email string) (*models.User, error) {
	query := `SELECT id, email, password, username, created_at FROM users WHERE email = ?`

	user := &models.User{}
	err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Username,
		&user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // user not found (not an error)
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

// FindByID finds a user by ID
func (r *SQLiteUserRepository) FindByID(id int) (*models.User, error) {
	query := `SELECT id, email, password, username, created_at FROM users WHERE id = ?`

	user := &models.User{}
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Username,
		&user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}
