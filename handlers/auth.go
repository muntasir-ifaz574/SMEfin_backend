package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"sme_fin_backend/models"
	"sme_fin_backend/utils"
)

type AuthHandler struct {
	DB *sql.DB
}

type SendOTPRequest struct {
	Email string `json:"email"`
}

type VerifyOTPRequest struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

type VerifyOTPResponse struct {
	Token         string `json:"token"`
	UserID        string `json:"user_id"`
	Email         string `json:"email"`
	AccountStatus string `json:"account_status"`
}

func (h *AuthHandler) SendOTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.SendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SendOTPRequest

	// Parse form-data or JSON
	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") || strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
		if strings.HasPrefix(contentType, "multipart/form-data") {
			if err := r.ParseMultipartForm(32 << 20); err != nil {
				utils.SendErrorResponse(w, "Invalid form data", http.StatusBadRequest)
				return
			}
			req.Email = r.FormValue("email")
		} else {
			if err := r.ParseForm(); err != nil {
				utils.SendErrorResponse(w, "Invalid form data", http.StatusBadRequest)
				return
			}
			req.Email = r.FormValue("email")
		}
	} else {
		// JSON fallback
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.SendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
			return
		}
	}

	// Validate email
	if req.Email == "" {
		utils.SendErrorResponse(w, "Email is required", http.StatusBadRequest)
		return
	}

	if !utils.ValidateEmail(req.Email) {
		utils.SendErrorResponse(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	// Get default OTP from environment
	defaultOTP := os.Getenv("DEFAULT_OTP")
	if defaultOTP == "" {
		defaultOTP = "123456"
	}

	// Create or get user
	user, err := models.GetUserByEmail(h.DB, req.Email)
	if err != nil {
		utils.SendErrorResponse(w, "Database error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		user = &models.User{Email: req.Email}
		if err := user.Create(h.DB); err != nil {
			utils.SendErrorResponse(w, "Failed to create user", http.StatusInternalServerError)
			return
		}
	}

	// Create OTP verification record
	otpVerification := &models.OTPVerification{
		Email: req.Email,
		OTP:   defaultOTP,
	}

	if err := otpVerification.Create(h.DB); err != nil {
		utils.SendErrorResponse(w, "Failed to create OTP verification", http.StatusInternalServerError)
		return
	}

	// In production, send OTP via email/SMS
	// For now, we'll just return success

	utils.SendSuccessResponse(w, "OTP sent successfully", map[string]string{
		"email":   req.Email,
		"message": "OTP sent to email (use default OTP: " + defaultOTP + " for testing)",
	}, http.StatusOK)
}

func (h *AuthHandler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.SendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req VerifyOTPRequest

	// Parse form-data or JSON
	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") || strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
		if strings.HasPrefix(contentType, "multipart/form-data") {
			if err := r.ParseMultipartForm(32 << 20); err != nil {
				utils.SendErrorResponse(w, "Invalid form data", http.StatusBadRequest)
				return
			}
			req.Email = r.FormValue("email")
			req.OTP = r.FormValue("otp")
		} else {
			if err := r.ParseForm(); err != nil {
				utils.SendErrorResponse(w, "Invalid form data", http.StatusBadRequest)
				return
			}
			req.Email = r.FormValue("email")
			req.OTP = r.FormValue("otp")
		}
	} else {
		// JSON fallback
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.SendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
			return
		}
	}

	// Validate inputs
	if req.Email == "" {
		utils.SendErrorResponse(w, "Email is required", http.StatusBadRequest)
		return
	}

	if req.OTP == "" {
		utils.SendErrorResponse(w, "OTP is required", http.StatusBadRequest)
		return
	}

	if !utils.ValidateOTP(req.OTP) {
		utils.SendErrorResponse(w, "Invalid OTP format", http.StatusBadRequest)
		return
	}

	// Verify OTP
	otpVerification, err := models.VerifyOTP(h.DB, req.Email, req.OTP)
	if err != nil {
		utils.SendErrorResponse(w, "Database error", http.StatusInternalServerError)
		return
	}

	if otpVerification == nil {
		utils.SendErrorResponse(w, "Invalid or expired OTP", http.StatusUnauthorized)
		return
	}

	// Get user
	user, err := models.GetUserByEmail(h.DB, req.Email)
	if err != nil {
		utils.SendErrorResponse(w, "Database error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		utils.SendErrorResponse(w, "User not found", http.StatusNotFound)
		return
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user.ID, user.Email)
	if err != nil {
		utils.SendErrorResponse(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Get account status
	accountStatus, err := models.GetAccountStatus(h.DB, user.ID)
	if err != nil {
		utils.SendErrorResponse(w, "Failed to get account status", http.StatusInternalServerError)
		return
	}

	response := VerifyOTPResponse{
		Token:         token,
		UserID:        user.ID.String(),
		Email:         user.Email,
		AccountStatus: accountStatus.Status,
	}

	utils.SendSuccessResponse(w, "OTP verified successfully", response, http.StatusOK)
}
