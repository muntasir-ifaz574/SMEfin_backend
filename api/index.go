package handler

import (
	"database/sql"
	"log"
	"net/http"
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

func dbOrError(w http.ResponseWriter) *sql.DB {
	d := getDB()
	if d == nil {
		utils.SendErrorResponse(w, "Database not configured. Check DB_* env vars.", http.StatusInternalServerError)
	}
	return d
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
		protected.HandleFunc("/user/personal-details", func(w http.ResponseWriter, r *http.Request) {
			d := dbOrError(w)
			if d == nil {
				return
			}
			(&handlers.UserHandler{DB: d}).PersonalDetails(w, r)
		}).Methods("POST")
		protected.HandleFunc("/user/business-details", func(w http.ResponseWriter, r *http.Request) {
			d := dbOrError(w)
			if d == nil {
				return
			}
			(&handlers.UserHandler{DB: d}).BusinessDetails(w, r)
		}).Methods("POST")
		protected.HandleFunc("/user/trade-license", func(w http.ResponseWriter, r *http.Request) {
			d := dbOrError(w)
			if d == nil {
				return
			}
			(&handlers.UserHandler{DB: d}).TradeLicense(w, r)
		}).Methods("POST")
		protected.HandleFunc("/user/full-registration", func(w http.ResponseWriter, r *http.Request) {
			d := dbOrError(w)
			if d == nil {
				return
			}
			(&handlers.UserHandler{DB: d}).FullRegistration(w, r)
		}).Methods("POST")
		protected.HandleFunc("/user/submit", func(w http.ResponseWriter, r *http.Request) {
			d := dbOrError(w)
			if d == nil {
				return
			}
			(&handlers.UserHandler{DB: d}).Submit(w, r)
		}).Methods("POST")
		protected.HandleFunc("/user/status", func(w http.ResponseWriter, r *http.Request) {
			d := dbOrError(w)
			if d == nil {
				return
			}
			(&handlers.UserHandler{DB: d}).Status(w, r)
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
