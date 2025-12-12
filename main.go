package main

import (
	"log"
	"net/http"
	"os"

	"sme_fin_backend/database"
	"sme_fin_backend/handlers"
	"sme_fin_backend/middleware"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Connect to database
	db, err := database.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize handlers
	authHandler := &handlers.AuthHandler{DB: db}
	userHandler := &handlers.UserHandler{DB: db}

	// Setup router
	router := mux.NewRouter()

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","message":"Server is running"}`))
	}).Methods("GET")

	// Public routes
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/auth/send-otp", authHandler.SendOTP).Methods("POST")
	api.HandleFunc("/auth/verify-otp", authHandler.VerifyOTP).Methods("POST")

	// Protected routes
	protected := api.PathPrefix("").Subrouter()
	protected.Use(middleware.JWTAuthMiddleware)
	protected.HandleFunc("/user/personal-details", userHandler.PersonalDetails).Methods("POST")
	protected.HandleFunc("/user/business-details", userHandler.BusinessDetails).Methods("POST")
	protected.HandleFunc("/user/trade-license", userHandler.TradeLicense).Methods("POST")
	protected.HandleFunc("/user/full-registration", userHandler.FullRegistration).Methods("POST")
	protected.HandleFunc("/user/submit", userHandler.Submit).Methods("POST")
	protected.HandleFunc("/user/status", userHandler.Status).Methods("GET")

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

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, corsHandler(router)))
}
