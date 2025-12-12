package handler

import (
	"database/sql"
	"log"
	"net/http"
	"sync"

	"sme_fin_backend/database"
	"sme_fin_backend/handlers"
	"sme_fin_backend/middleware"

	"github.com/gorilla/mux"
)

var (
	router     *mux.Router
	db         *sql.DB
	dbOnce     sync.Once
	routerOnce sync.Once
)

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
		protected.HandleFunc("/user/personal-details", func(w http.ResponseWriter, r *http.Request) {
			(&handlers.UserHandler{DB: getDB()}).PersonalDetails(w, r)
		}).Methods("POST")
		protected.HandleFunc("/user/business-details", func(w http.ResponseWriter, r *http.Request) {
			(&handlers.UserHandler{DB: getDB()}).BusinessDetails(w, r)
		}).Methods("POST")
		protected.HandleFunc("/user/trade-license", func(w http.ResponseWriter, r *http.Request) {
			(&handlers.UserHandler{DB: getDB()}).TradeLicense(w, r)
		}).Methods("POST")
		protected.HandleFunc("/user/full-registration", func(w http.ResponseWriter, r *http.Request) {
			(&handlers.UserHandler{DB: getDB()}).FullRegistration(w, r)
		}).Methods("POST")
		protected.HandleFunc("/user/submit", func(w http.ResponseWriter, r *http.Request) {
			(&handlers.UserHandler{DB: getDB()}).Submit(w, r)
		}).Methods("POST")
		protected.HandleFunc("/user/status", func(w http.ResponseWriter, r *http.Request) {
			(&handlers.UserHandler{DB: getDB()}).Status(w, r)
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

