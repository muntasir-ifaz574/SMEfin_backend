package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

func Connect() (*sql.DB, error) {
	var connStr string

	// Prefer a single DATABASE_URL if provided (works well with Supabase / Vercel)
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL != "" {
		connStr = databaseURL
	} else {
		// Build connection string from individual env vars
		host := os.Getenv("DB_HOST")
		port := os.Getenv("DB_PORT")
		user := os.Getenv("DB_USER")
		password := os.Getenv("DB_PASSWORD")
		dbname := os.Getenv("DB_NAME")
		sslmode := os.Getenv("DB_SSLMODE")

		if sslmode == "" {
			sslmode = "require"
		}

		// Validate required env vars
		if host == "" || port == "" || user == "" || password == "" || dbname == "" {
			return nil, fmt.Errorf("missing required database env vars (DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME) or DATABASE_URL")
		}

		connStr = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			host, port, user, password, dbname, sslmode)
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings for serverless environments
	// Use smaller pool sizes for serverless to avoid connection limits
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(0) // Reuse connections

	// Don't ping immediately in serverless - connections are lazy
	// The first query will establish the connection
	// This avoids cold start issues in serverless environments

	return db, nil
}
