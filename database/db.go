package database

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

// getEnv gets environment variable with fallback options
func getEnv(keys ...string) string {
	for _, key := range keys {
		if value := os.Getenv(key); value != "" {
			return value
		}
	}
	return ""
}

func Connect() (*sql.DB, error) {
	var connStr string

	// Try multiple possible environment variable names for database URL
	// Common names: DATABASE_URL, POSTGRES_URL, POSTGRES_PRISMA_URL, POSTGRES_URL_NON_POOLING
	databaseURL := getEnv("DATABASE_URL", "POSTGRES_URL", "POSTGRES_PRISMA_URL", "POSTGRES_URL_NON_POOLING")

	if databaseURL != "" {
		// If DATABASE_URL doesn't have sslmode, add it for Supabase
		if !strings.Contains(databaseURL, "sslmode=") {
			if strings.Contains(databaseURL, "?") {
				databaseURL += "&sslmode=require"
			} else {
				databaseURL += "?sslmode=require"
			}
		}
		connStr = databaseURL
	} else {
		// Try multiple naming conventions for individual env vars
		host := getEnv("DB_HOST", "POSTGRES_HOST", "PGHOST")
		port := getEnv("DB_PORT", "POSTGRES_PORT", "PGPORT")
		user := getEnv("DB_USER", "POSTGRES_USER", "PGUSER", "POSTGRES_USERNAME")
		password := getEnv("DB_PASSWORD", "POSTGRES_PASSWORD", "PGPASSWORD")
		dbname := getEnv("DB_NAME", "POSTGRES_DATABASE", "POSTGRES_DB", "PGDATABASE")
		sslmode := getEnv("DB_SSLMODE", "POSTGRES_SSLMODE", "PGSSLMODE")

		if sslmode == "" {
			sslmode = "require"
		}

		// Default port if not provided
		if port == "" {
			port = "5432"
		}

		// Validate required env vars
		if host == "" || user == "" || password == "" || dbname == "" {
			// Log available env vars for debugging (without sensitive data)
			availableVars := []string{}
			for _, key := range []string{"DATABASE_URL", "POSTGRES_URL", "DB_HOST", "POSTGRES_HOST", "DB_USER", "POSTGRES_USER", "DB_NAME", "POSTGRES_DATABASE"} {
				if os.Getenv(key) != "" {
					availableVars = append(availableVars, key)
				}
			}

			return nil, fmt.Errorf("missing required database env vars. Need either DATABASE_URL/POSTGRES_URL, or (DB_HOST/POSTGRES_HOST, DB_USER/POSTGRES_USER, DB_PASSWORD/POSTGRES_PASSWORD, DB_NAME/POSTGRES_DATABASE). Found env vars: %v", availableVars)
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
