package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"sme_fin_backend/models"
	"sme_fin_backend/storage"
	"sme_fin_backend/utils"

	"github.com/google/uuid"
)

type UserHandler struct {
	DB *sql.DB
}

type PersonalDetailsRequest struct {
	FullName    string `json:"full_name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
}

type BusinessDetailsRequest struct {
	BusinessName       string `json:"business_name"`
	TradeLicenseNumber string `json:"trade_license_number"`
}

type TradeLicenseRequest struct {
	Filename string `json:"filename"`
	FileURL  string `json:"file_url"`
}

// FullRegistrationRequest groups all onboarding data into a single payload.
type FullRegistrationRequest struct {
	Personal PersonalDetailsRequest `json:"personal"`
	Business BusinessDetailsRequest `json:"business"`
	Trade    TradeLicenseRequest    `json:"trade"`
}

func (h *UserHandler) getUserIDFromRequest(r *http.Request) (uuid.UUID, error) {
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		return uuid.Nil, nil
	}
	return uuid.Parse(userIDStr)
}

// getFormValue tries multiple form field names and returns the first non-empty value
func getFormValue(r *http.Request, keys ...string) string {
	for _, key := range keys {
		if value := r.FormValue(key); value != "" {
			return value
		}
	}
	return ""
}

// GetUserData retrieves all user registration data
func (h *UserHandler) GetUserData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.SendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := h.getUserIDFromRequest(r)
	if err != nil || userID == uuid.Nil {
		utils.SendErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user
	user, err := models.GetUserByID(h.DB, userID)
	if err != nil {
		utils.SendErrorResponse(w, "Database error", http.StatusInternalServerError)
		return
	}
	if user == nil {
		utils.SendErrorResponse(w, "User not found", http.StatusNotFound)
		return
	}

	// Get personal details
	personalDetails, err := models.GetPersonalDetails(h.DB, userID)
	if err != nil {
		utils.SendErrorResponse(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Get business details
	businessDetails, err := models.GetBusinessDetails(h.DB, userID)
	if err != nil {
		utils.SendErrorResponse(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Get trade license
	tradeLicense, err := models.GetTradeLicense(h.DB, userID)
	if err != nil {
		utils.SendErrorResponse(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Get account status
	accountStatus, err := models.GetAccountStatus(h.DB, userID)
	if err != nil {
		utils.SendErrorResponse(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Build response
	response := map[string]interface{}{
		"user_id": userID.String(),
		"email":   user.Email,
		"status":  accountStatus.Status,
	}

	if personalDetails != nil {
		response["personal"] = personalDetails
	} else {
		response["personal"] = nil
	}

	if businessDetails != nil {
		response["business"] = businessDetails
	} else {
		response["business"] = nil
	}

	if tradeLicense != nil {
		response["trade_license"] = tradeLicense
	} else {
		response["trade_license"] = nil
	}

	utils.SendSuccessResponse(w, "User data retrieved successfully", response, http.StatusOK)
}

func (h *UserHandler) Status(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.SendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := h.getUserIDFromRequest(r)
	if err != nil || userID == uuid.Nil {
		utils.SendErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	accountStatus, err := models.GetAccountStatus(h.DB, userID)
	if err != nil {
		utils.SendErrorResponse(w, "Database error", http.StatusInternalServerError)
		return
	}

	if accountStatus == nil {
		utils.SendErrorResponse(w, "Account not found", http.StatusNotFound)
		return
	}

	utils.SendSuccessResponse(w, "Account status retrieved successfully", accountStatus, http.StatusOK)
}

// FullRegistration handles personal, business, and trade license in a single API call.
func (h *UserHandler) FullRegistration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.SendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := h.getUserIDFromRequest(r)
	if err != nil || userID == uuid.Nil {
		utils.SendErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req FullRegistrationRequest

	// Parse form-data or JSON
	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") || strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
		if strings.HasPrefix(contentType, "multipart/form-data") {
			if err := r.ParseMultipartForm(32 << 20); err != nil {
				utils.SendErrorResponse(w, "Invalid form data", http.StatusBadRequest)
				return
			}
		} else {
			if err := r.ParseForm(); err != nil {
				utils.SendErrorResponse(w, "Invalid form data", http.StatusBadRequest)
				return
			}
		}

		// Parse form fields - support both nested (personal[full_name]) and flat (personal_full_name) formats
		req.Personal.FullName = getFormValue(r, "personal[full_name]", "personal_full_name", "full_name")
		req.Personal.Email = getFormValue(r, "personal[email]", "personal_email", "email")
		req.Personal.PhoneNumber = getFormValue(r, "personal[phone_number]", "personal_phone_number", "phone_number")

		req.Business.BusinessName = getFormValue(r, "business[business_name]", "business_business_name", "business_name")
		req.Business.TradeLicenseNumber = getFormValue(r, "business[trade_license_number]", "business_trade_license_number", "trade_license_number")

		// Handle file upload for trade license if present
		if strings.HasPrefix(contentType, "multipart/form-data") {
			file, fileHeader, err := r.FormFile("trade[file]")
			if err == nil && file != nil {
				defer file.Close()
				
				// Validate file type (PDF, JPG, PNG)
				allowedTypes := []string{"pdf", "jpg", "jpeg", "png"}
				if !utils.ValidateFileType(fileHeader.Filename, allowedTypes) {
					utils.SendErrorResponse(w, "Invalid file type. Only PDF, JPG, and PNG files are allowed", http.StatusBadRequest)
					return
				}
				
				// Validate file size (max 10MB)
				maxSizeMB := 10
				if !utils.ValidateFileSize(fileHeader.Size, maxSizeMB) {
					utils.SendErrorResponse(w, fmt.Sprintf("File size exceeds %dMB limit", maxSizeMB), http.StatusBadRequest)
					return
				}
				
				req.Trade.Filename = fileHeader.Filename

				// Upload to Supabase storage
				bucketName := os.Getenv("SUPABASE_BUCKET_NAME")
				if bucketName == "" {
					bucketName = "vercel_bucket" // Default bucket name
				}

				fileURL, uploadErr := storage.UploadFileToSupabase(file, fileHeader.Filename, bucketName)
				if uploadErr != nil {
					log.Printf("Failed to upload file to Supabase: %v", uploadErr)
					utils.SendErrorResponse(w, fmt.Sprintf("Failed to upload file: %v", uploadErr), http.StatusInternalServerError)
					return
				}
				req.Trade.FileURL = fileURL
			} else {
				// Fallback to form values
				req.Trade.Filename = getFormValue(r, "trade[filename]", "trade_filename", "filename")
				req.Trade.FileURL = getFormValue(r, "trade[file_url]", "trade_file_url", "file_url")
			}
		} else {
			req.Trade.Filename = getFormValue(r, "trade[filename]", "trade_filename", "filename")
			req.Trade.FileURL = getFormValue(r, "trade[file_url]", "trade_file_url", "file_url")
		}
	} else {
		// JSON fallback
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.SendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
			return
		}
	}

	// Validate personal details
	if req.Personal.FullName == "" {
		utils.SendErrorResponse(w, "Full name is required", http.StatusBadRequest)
		return
	}
	if req.Personal.Email == "" {
		utils.SendErrorResponse(w, "Email is required", http.StatusBadRequest)
		return
	}
	if !utils.ValidateEmail(req.Personal.Email) {
		utils.SendErrorResponse(w, "Invalid email format", http.StatusBadRequest)
		return
	}
	if req.Personal.PhoneNumber == "" {
		utils.SendErrorResponse(w, "Phone number is required", http.StatusBadRequest)
		return
	}
	if !utils.ValidatePhone(req.Personal.PhoneNumber) {
		utils.SendErrorResponse(w, "Invalid phone number format", http.StatusBadRequest)
		return
	}

	// Validate business details
	if req.Business.BusinessName == "" {
		utils.SendErrorResponse(w, "Business name is required", http.StatusBadRequest)
		return
	}
	if req.Business.TradeLicenseNumber == "" {
		utils.SendErrorResponse(w, "Trade license number is required", http.StatusBadRequest)
		return
	}

	// Validate trade license
	if req.Trade.Filename == "" {
		utils.SendErrorResponse(w, "Filename is required", http.StatusBadRequest)
		return
	}
	if req.Trade.FileURL == "" {
		utils.SendErrorResponse(w, "File URL is required (or upload a file)", http.StatusBadRequest)
		return
	}

	// Persist personal details
	personalDetails := &models.PersonalDetails{
		UserID:      userID,
		FullName:    req.Personal.FullName,
		Email:       req.Personal.Email,
		PhoneNumber: req.Personal.PhoneNumber,
	}
	if err := personalDetails.CreateOrUpdate(h.DB); err != nil {
		utils.SendErrorResponse(w, "Failed to save personal details", http.StatusInternalServerError)
		return
	}

	// Persist business details
	businessDetails := &models.BusinessDetails{
		UserID:             userID,
		BusinessName:       req.Business.BusinessName,
		TradeLicenseNumber: req.Business.TradeLicenseNumber,
	}
	if err := businessDetails.CreateOrUpdate(h.DB); err != nil {
		utils.SendErrorResponse(w, "Failed to save business details", http.StatusInternalServerError)
		return
	}

	// Persist trade license
	tradeLicense := &models.TradeLicense{
		UserID:   userID,
		Filename: req.Trade.Filename,
		FileURL:  req.Trade.FileURL,
	}
	if err := tradeLicense.CreateOrUpdate(h.DB); err != nil {
		utils.SendErrorResponse(w, "Failed to save trade license", http.StatusInternalServerError)
		return
	}

	// Fetch status and summary
	accountStatus, err := models.GetAccountStatus(h.DB, userID)
	if err != nil {
		utils.SendErrorResponse(w, "Failed to get account status", http.StatusInternalServerError)
		return
	}

	summary, err := models.GetRegistrationSummary(h.DB, userID)
	if err != nil {
		utils.SendErrorResponse(w, "Failed to get registration summary", http.StatusInternalServerError)
		return
	}

	utils.SendSuccessResponse(w, "Full registration saved successfully", map[string]interface{}{
		"personal": personalDetails,
		"business": businessDetails,
		"trade":    tradeLicense,
		"status":   accountStatus.Status,
		"summary":  summary,
	}, http.StatusOK)
}
