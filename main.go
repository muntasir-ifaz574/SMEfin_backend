package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync"

	"sme_fin_backend/database"
	"sme_fin_backend/handlers"
	"sme_fin_backend/middleware"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

var (
	router     *mux.Router
	db         *sql.DB
	dbOnce     sync.Once
	routerOnce sync.Once
)

func init() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}
}

func getDB() *sql.DB {
	dbOnce.Do(func() {
		var err error
		db, err = database.Connect()
		if err != nil {
			log.Printf("Failed to connect to database: %v", err)
		}
	})
	return db
}

func getRouter() *mux.Router {
	routerOnce.Do(func() {
		router = mux.NewRouter()

		// Health check endpoint
		router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok","message":"Server is running"}`))
		}).Methods("GET")

		// Public routes
		api := router.PathPrefix("/api").Subrouter()
		api.HandleFunc("/auth/send-otp", func(w http.ResponseWriter, r *http.Request) {
			(&handlers.AuthHandler{DB: getDB()}).SendOTP(w, r)
		}).Methods("POST")
		api.HandleFunc("/auth/verify-otp", func(w http.ResponseWriter, r *http.Request) {
			(&handlers.AuthHandler{DB: getDB()}).VerifyOTP(w, r)
		}).Methods("POST")

		// Protected routes
		protected := api.PathPrefix("").Subrouter()
		protected.Use(middleware.JWTAuthMiddleware)
		protected.HandleFunc("/user/full-registration", func(w http.ResponseWriter, r *http.Request) {
			(&handlers.UserHandler{DB: getDB()}).FullRegistration(w, r)
		}).Methods("POST")
		protected.HandleFunc("/user/status", func(w http.ResponseWriter, r *http.Request) {
			(&handlers.UserHandler{DB: getDB()}).Status(w, r)
		}).Methods("GET")
		protected.HandleFunc("/user/data", func(w http.ResponseWriter, r *http.Request) {
			(&handlers.UserHandler{DB: getDB()}).GetUserData(w, r)
		}).Methods("GET")
		protected.HandleFunc("/financing/request", func(w http.ResponseWriter, r *http.Request) {
			(&handlers.FinancingHandler{DB: getDB()}).RequestFinancing(w, r)
		}).Methods("POST")
		protected.HandleFunc("/financing/requests", func(w http.ResponseWriter, r *http.Request) {
			(&handlers.FinancingHandler{DB: getDB()}).GetFinancingRequests(w, r)
		}).Methods("GET")
		protected.HandleFunc("/financing/request-detail", func(w http.ResponseWriter, r *http.Request) {
			(&handlers.FinancingHandler{DB: getDB()}).GetFinancingRequest(w, r)
		}).Methods("GET")
		protected.HandleFunc("/financing/latest", func(w http.ResponseWriter, r *http.Request) {
			(&handlers.FinancingHandler{DB: getDB()}).GetLatestFinancingRequest(w, r)
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

// main function for local development
func main() {
	// Connect to database for local development
	db := getDB()
	if db == nil {
		log.Fatal("Failed to connect to database")
	}
	defer db.Close()

	// Initialize router
	r := getRouter()

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
