package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/simplejourney/helpdesk/internal/models"
)

var ErrNotFound = errors.New("not found")
var ErrConflict = errors.New("conflict")

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(name, email, passwordHash string, role models.Role) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRow(`
		INSERT INTO users (name, email, password_hash, role)
		VALUES ($1,$2,$3,$4)
		RETURNING id, name, email, password_hash, role, created_at
	`, name, email, passwordHash, role).Scan(
		&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrConflict
		}
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRow(`
		SELECT id, name, email, password_hash, role, created_at
		FROM users WHERE email=$1
	`, email).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return user, err
}

func (r *UserRepository) FindByID(id string) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRow(`
		SELECT id, name, email, password_hash, role, created_at
		FROM users WHERE id=$1
	`, id).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return user, err
}

func (r *UserRepository) List() ([]models.User, error) {
	rows, err := r.db.Query(`
		SELECT id, name, email, password_hash, role, created_at
		FROM users ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *UserRepository) UpdateRole(id string, role models.Role) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRow(`
		UPDATE users SET role=$2 WHERE id=$1
		RETURNING id, name, email, password_hash, role, created_at
	`, id, role).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return user, err
}

func isUniqueViolation(err error) bool {
	return err != nil && (contains(err.Error(), "duplicate key") || contains(err.Error(), "unique constraint"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(func() bool {
			for i := 0; i+len(substr) <= len(s); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		})())
}

func BriefFromUser(u *models.User) *models.UserBrief {
	if u == nil {
		return nil
	}
	return &models.UserBrief{ID: u.ID, Name: u.Name, Email: u.Email, Role: u.Role}
}

func FormatPage(total, page, perPage int) (int, int) {
	if perPage <= 0 {
		perPage = 10
	}
	if page <= 0 {
		page = 1
	}
	totalPages := total / perPage
	if total%perPage != 0 {
		totalPages++
	}
	if totalPages == 0 {
		totalPages = 1
	}
	return page, totalPages
}

func SafeSort(sortBy, sortDir string) (string, string) {
	cols := map[string]string{
		"created_at": "t.created_at",
		"updated_at": "t.updated_at",
		"priority":   "t.priority",
		"status":     "t.status",
		"title":      "t.title",
	}
	col, ok := cols[sortBy]
	if !ok {
		col = "t.created_at"
	}
	dir := "DESC"
	if sortDir == "asc" {
		dir = "ASC"
	}
	return col, dir
}

func PaginateOffset(page, perPage int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 10
	}
	if perPage > 100 {
		perPage = 100
	}
	return (page - 1) * perPage, perPage
}

func NullString(ptr *string) sql.NullString {
	if ptr == nil || *ptr == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: *ptr, Valid: true}
}

func PtrString(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	s := ns.String
	return &s
}

func Wrap(err error, msg string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", msg, err)
}
