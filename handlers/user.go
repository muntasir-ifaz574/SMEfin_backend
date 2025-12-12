package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"sme_fin_backend/models"
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

func (h *UserHandler) PersonalDetails(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.SendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := h.getUserIDFromRequest(r)
	if err != nil || userID == uuid.Nil {
		utils.SendErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req PersonalDetailsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate inputs
	if req.FullName == "" {
		utils.SendErrorResponse(w, "Full name is required", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		utils.SendErrorResponse(w, "Email is required", http.StatusBadRequest)
		return
	}

	if !utils.ValidateEmail(req.Email) {
		utils.SendErrorResponse(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	if req.PhoneNumber == "" {
		utils.SendErrorResponse(w, "Phone number is required", http.StatusBadRequest)
		return
	}

	if !utils.ValidatePhone(req.PhoneNumber) {
		utils.SendErrorResponse(w, "Invalid phone number format", http.StatusBadRequest)
		return
	}

	// Create or update personal details
	personalDetails := &models.PersonalDetails{
		UserID:      userID,
		FullName:    req.FullName,
		Email:       req.Email,
		PhoneNumber: req.PhoneNumber,
	}

	if err := personalDetails.CreateOrUpdate(h.DB); err != nil {
		utils.SendErrorResponse(w, "Failed to save personal details", http.StatusInternalServerError)
		return
	}

	utils.SendSuccessResponse(w, "Personal details saved successfully", personalDetails, http.StatusOK)
}

func (h *UserHandler) BusinessDetails(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.SendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := h.getUserIDFromRequest(r)
	if err != nil || userID == uuid.Nil {
		utils.SendErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req BusinessDetailsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate inputs
	if req.BusinessName == "" {
		utils.SendErrorResponse(w, "Business name is required", http.StatusBadRequest)
		return
	}

	if req.TradeLicenseNumber == "" {
		utils.SendErrorResponse(w, "Trade license number is required", http.StatusBadRequest)
		return
	}

	// Create or update business details
	businessDetails := &models.BusinessDetails{
		UserID:             userID,
		BusinessName:       req.BusinessName,
		TradeLicenseNumber: req.TradeLicenseNumber,
	}

	if err := businessDetails.CreateOrUpdate(h.DB); err != nil {
		utils.SendErrorResponse(w, "Failed to save business details", http.StatusInternalServerError)
		return
	}

	utils.SendSuccessResponse(w, "Business details saved successfully", businessDetails, http.StatusOK)
}

func (h *UserHandler) TradeLicense(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.SendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := h.getUserIDFromRequest(r)
	if err != nil || userID == uuid.Nil {
		utils.SendErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req TradeLicenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate inputs
	if req.Filename == "" {
		utils.SendErrorResponse(w, "Filename is required", http.StatusBadRequest)
		return
	}

	if req.FileURL == "" {
		utils.SendErrorResponse(w, "File URL is required", http.StatusBadRequest)
		return
	}

	// Create or update trade license
	tradeLicense := &models.TradeLicense{
		UserID:   userID,
		Filename: req.Filename,
		FileURL:  req.FileURL,
	}

	if err := tradeLicense.CreateOrUpdate(h.DB); err != nil {
		utils.SendErrorResponse(w, "Failed to save trade license", http.StatusInternalServerError)
		return
	}

	utils.SendSuccessResponse(w, "Trade license saved successfully", tradeLicense, http.StatusOK)
}

func (h *UserHandler) Submit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.SendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := h.getUserIDFromRequest(r)
	if err != nil || userID == uuid.Nil {
		utils.SendErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get registration summary
	summary, err := models.GetRegistrationSummary(h.DB, userID)
	if err != nil {
		utils.SendErrorResponse(w, "Database error", http.StatusInternalServerError)
		return
	}

	if summary == nil {
		utils.SendErrorResponse(w, "Registration incomplete. Please complete all steps.", http.StatusBadRequest)
		return
	}

	// Update account status to "old" (complete)
	accountStatus, err := models.GetAccountStatus(h.DB, userID)
	if err != nil {
		utils.SendErrorResponse(w, "Failed to get account status", http.StatusInternalServerError)
		return
	}

	utils.SendSuccessResponse(w, "Registration submitted successfully", map[string]interface{}{
		"summary": summary,
		"status":  accountStatus.Status,
		"message": "Your details are submitted.",
	}, http.StatusOK)
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
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
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
		utils.SendErrorResponse(w, "File URL is required", http.StatusBadRequest)
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
