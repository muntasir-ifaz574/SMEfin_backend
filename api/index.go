package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"sme_fin_backend/database"
	"sme_fin_backend/handlers"
	"sme_fin_backend/middleware"
	"sme_fin_backend/utils"

	"github.com/gorilla/mux"
)

var (
	router     *mux.Router
	db         *sql.DB
	dbOnce     sync.Once
	routerOnce sync.Once
)

func getDB() (*sql.DB, error) {
	var err error
	dbOnce.Do(func() {
		db, err = database.Connect()
		if err != nil {
			log.Printf("Failed to connect to database: %v", err)
		}
	})
	if err != nil {
		return nil, err
	}
	if db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}
	return db, nil
}

func dbOrError(w http.ResponseWriter) *sql.DB {
	d, err := getDB()
	if err != nil {
		log.Printf("Database error: %v", err)
		utils.SendErrorResponse(w, fmt.Sprintf("Database connection error: %v", err), http.StatusInternalServerError)
		return nil
	}
	if d == nil {
		utils.SendErrorResponse(w, "Database connection is not available", http.StatusInternalServerError)
		return nil
	}
	return d
}

func getRouter() *mux.Router {
	routerOnce.Do(func() {
		router = mux.NewRouter()

		// Health check endpoint
		router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			// Check database connection
			db, err := getDB()
			dbStatus := "connected"
			if err != nil || db == nil {
				dbStatus = "disconnected"
			}

			response := fmt.Sprintf(`{"status":"ok","message":"Server is running","database":"%s"}`, dbStatus)
			w.Write([]byte(response))
		}).Methods("GET")

		// Debug endpoint to check env vars (without sensitive data)
		router.HandleFunc("/debug/env", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			envVars := make(map[string]string)
			checkVars := []string{
				"DATABASE_URL", "POSTGRES_URL", "POSTGRES_PRISMA_URL",
				"DB_HOST", "POSTGRES_HOST", "PGHOST",
				"DB_PORT", "POSTGRES_PORT", "PGPORT",
				"DB_USER", "POSTGRES_USER", "PGUSER",
				"DB_PASSWORD", "POSTGRES_PASSWORD", "PGPASSWORD",
				"DB_NAME", "POSTGRES_DATABASE", "POSTGRES_DB", "PGDATABASE",
				"DB_SSLMODE", "POSTGRES_SSLMODE",
			}

			for _, key := range checkVars {
				if value := os.Getenv(key); value != "" {
					// Mask passwords
					if strings.Contains(strings.ToLower(key), "password") {
						envVars[key] = "***masked***"
					} else {
						envVars[key] = value
					}
				}
			}

			json.NewEncoder(w).Encode(map[string]interface{}{
				"available_env_vars": envVars,
				"total_found":        len(envVars),
			})
		}).Methods("GET")

		// Public routes
		api := router.PathPrefix("/api").Subrouter()
		api.HandleFunc("/auth/send-otp", func(w http.ResponseWriter, r *http.Request) {
			d := dbOrError(w)
			if d == nil {
				return
			}
			(&handlers.AuthHandler{DB: d}).SendOTP(w, r)
		}).Methods("POST")
		api.HandleFunc("/auth/verify-otp", func(w http.ResponseWriter, r *http.Request) {
			d := dbOrError(w)
			if d == nil {
				return
			}
			(&handlers.AuthHandler{DB: d}).VerifyOTP(w, r)
		}).Methods("POST")

		// Protected routes
		protected := api.PathPrefix("").Subrouter()
		protected.Use(middleware.JWTAuthMiddleware)
		protected.HandleFunc("/user/full-registration", func(w http.ResponseWriter, r *http.Request) {
			d := dbOrError(w)
			if d == nil {
				return
			}
			(&handlers.UserHandler{DB: d}).FullRegistration(w, r)
		}).Methods("POST")
		protected.HandleFunc("/user/status", func(w http.ResponseWriter, r *http.Request) {
			d := dbOrError(w)
			if d == nil {
				return
			}
			(&handlers.UserHandler{DB: d}).Status(w, r)
		}).Methods("GET")
		protected.HandleFunc("/user/data", func(w http.ResponseWriter, r *http.Request) {
			d := dbOrError(w)
			if d == nil {
				return
			}
			(&handlers.UserHandler{DB: d}).GetUserData(w, r)
		}).Methods("GET")
		protected.HandleFunc("/financing/request", func(w http.ResponseWriter, r *http.Request) {
			d := dbOrError(w)
			if d == nil {
				return
			}
			(&handlers.FinancingHandler{DB: d}).RequestFinancing(w, r)
		}).Methods("POST")
		protected.HandleFunc("/financing/requests", func(w http.ResponseWriter, r *http.Request) {
			d := dbOrError(w)
			if d == nil {
				return
			}
			(&handlers.FinancingHandler{DB: d}).GetFinancingRequests(w, r)
		}).Methods("GET")
		protected.HandleFunc("/financing/request-detail", func(w http.ResponseWriter, r *http.Request) {
			d := dbOrError(w)
			if d == nil {
				return
			}
			(&handlers.FinancingHandler{DB: d}).GetFinancingRequest(w, r)
		}).Methods("GET")
		protected.HandleFunc("/financing/latest", func(w http.ResponseWriter, r *http.Request) {
			d := dbOrError(w)
			if d == nil {
				return
			}
			(&handlers.FinancingHandler{DB: d}).GetLatestFinancingRequest(w, r)
		}).Methods("GET")

		// CORS middleware
		corsHandler := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

				if r.Method == "OPTIONS" {
					w.WriteHeader(http.StatusOK)
					return
				}

				next.ServeHTTP(w, r)
			})
		}

		router.Use(corsHandler)
	})
	return router
}

// Handler is the entry point for Vercel serverless functions
// This function must be exported for Vercel to detect it
func Handler(w http.ResponseWriter, r *http.Request) {
	getRouter().ServeHTTP(w, r)
}
