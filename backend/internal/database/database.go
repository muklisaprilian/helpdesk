package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func Connect(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	for i := 0; i < 30; i++ {
		if err = db.Ping(); err == nil {
			return db, nil
		}
		time.Sleep(time.Second)
	}
	return nil, fmt.Errorf("database ping failed: %w", err)
}

func Migrate(db *sql.DB, migrationsDir string) error {
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			filename TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	var ups []string
	for _, e := range entries {
		name := e.Name()
		if strings.HasSuffix(name, ".up.sql") {
			ups = append(ups, name)
		}
	}
	sort.Strings(ups)

	for _, name := range ups {
		var exists bool
		if err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE filename=$1)`, name).Scan(&exists); err != nil {
			return err
		}
		if exists {
			continue
		}

		body, err := os.ReadFile(filepath.Join(migrationsDir, name))
		if err != nil {
			return err
		}

		tx, err := db.Begin()
		if err != nil {
			return err
		}
		if _, err := tx.Exec(string(body)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("apply %s: %w", name, err)
		}
		if _, err := tx.Exec(`INSERT INTO schema_migrations (filename) VALUES ($1)`, name); err != nil {
			_ = tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}

func Seed(db *sql.DB) error {
	seeds := []struct {
		name     string
		email    string
		password string
		role     string
	}{
		{"Admin User", "admin@helpdesk.local", "Admin123!", "admin"},
		{"Agent User", "agent@helpdesk.local", "Agent123!", "agent"},
		{"Requester User", "requester@helpdesk.local", "Requester123!", "requester"},
	}

	for _, s := range seeds {
		var exists bool
		if err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE email=$1)`, s.email).Scan(&exists); err != nil {
			return err
		}
		if exists {
			continue
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(s.password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		if _, err := db.Exec(
			`INSERT INTO users (name, email, password_hash, role) VALUES ($1,$2,$3,$4)`,
			s.name, s.email, string(hash), s.role,
		); err != nil {
			return err
		}
	}

	var ticketCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM tickets`).Scan(&ticketCount); err != nil {
		return err
	}
	if ticketCount == 0 {
		var requesterID string
		if err := db.QueryRow(`SELECT id FROM users WHERE email='requester@helpdesk.local'`).Scan(&requesterID); err != nil {
			return err
		}
		_, err := db.Exec(`
			INSERT INTO tickets (title, description, category, priority, status, requester_id)
			VALUES
			('Cannot access VPN', 'VPN client fails after company password reset.', 'Network', 'high', 'open', $1),
			('Laptop keyboard sticky', 'Several keys stick after a coffee spill.', 'Hardware', 'medium', 'open', $1)
		`, requesterID)
		if err != nil {
			return err
		}
	}
	return nil
}
