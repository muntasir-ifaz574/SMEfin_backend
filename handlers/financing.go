package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"sme_fin_backend/models"
	"sme_fin_backend/utils"

	"github.com/google/uuid"
)

type FinancingHandler struct {
	DB *sql.DB
}

type FinancingRequestRequest struct {
	Amount          string `json:"amount"`
	Purpose         string `json:"purpose"`
	RepaymentPeriod string `json:"repayment_period"`
}

func (h *FinancingHandler) getUserIDFromRequest(r *http.Request) (uuid.UUID, error) {
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		return uuid.Nil, nil
	}
	return uuid.Parse(userIDStr)
}

// RequestFinancing creates a new financing request
func (h *FinancingHandler) RequestFinancing(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.SendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := h.getUserIDFromRequest(r)
	if err != nil || userID == uuid.Nil {
		utils.SendErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if user has completed registration
	accountStatus, err := models.GetAccountStatus(h.DB, userID)
	if err != nil {
		utils.SendErrorResponse(w, "Database error", http.StatusInternalServerError)
		return
	}
	if accountStatus == nil || !accountStatus.IsComplete {
		utils.SendErrorResponse(w, "Please complete your registration before requesting financing", http.StatusBadRequest)
		return
	}

	var req FinancingRequestRequest

	// Parse form-data or JSON
	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") || strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
		if strings.HasPrefix(contentType, "multipart/form-data") {
			if err := r.ParseMultipartForm(32 << 20); err != nil {
				utils.SendErrorResponse(w, "Invalid form data", http.StatusBadRequest)
				return
			}
			req.Amount = r.FormValue("amount")
			req.Purpose = r.FormValue("purpose")
			req.RepaymentPeriod = r.FormValue("repayment_period")
		} else {
			if err := r.ParseForm(); err != nil {
				utils.SendErrorResponse(w, "Invalid form data", http.StatusBadRequest)
				return
			}
			req.Amount = r.FormValue("amount")
			req.Purpose = r.FormValue("purpose")
			req.RepaymentPeriod = r.FormValue("repayment_period")
		}
	} else {
		// JSON fallback
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.SendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
			return
		}
	}

	// Validate inputs
	if req.Amount == "" {
		utils.SendErrorResponse(w, "Amount is required", http.StatusBadRequest)
		return
	}
	amount, err := strconv.ParseFloat(req.Amount, 64)
	if err != nil || amount <= 0 {
		utils.SendErrorResponse(w, "Invalid amount. Must be a positive number", http.StatusBadRequest)
		return
	}

	if req.Purpose == "" {
		utils.SendErrorResponse(w, "Purpose is required", http.StatusBadRequest)
		return
	}

	if req.RepaymentPeriod == "" {
		utils.SendErrorResponse(w, "Repayment period is required", http.StatusBadRequest)
		return
	}
	repaymentPeriod, err := strconv.Atoi(req.RepaymentPeriod)
	if err != nil || repaymentPeriod <= 0 {
		utils.SendErrorResponse(w, "Invalid repayment period. Must be a positive number of months", http.StatusBadRequest)
		return
	}

	// Create financing request
	financingRequest := &models.FinancingRequest{
		UserID:          userID,
		Amount:          amount,
		Purpose:         req.Purpose,
		RepaymentPeriod: repaymentPeriod,
		Status:          "pending",
	}

	if err := financingRequest.Create(h.DB); err != nil {
		utils.SendErrorResponse(w, "Failed to create financing request", http.StatusInternalServerError)
		return
	}

	utils.SendSuccessResponse(w, "Financing request submitted successfully", financingRequest, http.StatusCreated)
}

// GetFinancingRequests retrieves all financing requests for the authenticated user
func (h *FinancingHandler) GetFinancingRequests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.SendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := h.getUserIDFromRequest(r)
	if err != nil || userID == uuid.Nil {
		utils.SendErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	requests, err := models.GetFinancingRequestsByUserID(h.DB, userID)
	if err != nil {
		utils.SendErrorResponse(w, "Database error", http.StatusInternalServerError)
		return
	}

	utils.SendSuccessResponse(w, "Financing requests retrieved successfully", requests, http.StatusOK)
}

// GetFinancingRequest retrieves a specific financing request by ID
func (h *FinancingHandler) GetFinancingRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.SendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := h.getUserIDFromRequest(r)
	if err != nil || userID == uuid.Nil {
		utils.SendErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get request ID from URL path
	requestIDStr := r.URL.Query().Get("id")
	if requestIDStr == "" {
		utils.SendErrorResponse(w, "Request ID is required", http.StatusBadRequest)
		return
	}

	requestID, err := uuid.Parse(requestIDStr)
	if err != nil {
		utils.SendErrorResponse(w, "Invalid request ID", http.StatusBadRequest)
		return
	}

	request, err := models.GetFinancingRequestByID(h.DB, requestID)
	if err != nil {
		utils.SendErrorResponse(w, "Database error", http.StatusInternalServerError)
		return
	}

	if request == nil {
		utils.SendErrorResponse(w, "Financing request not found", http.StatusNotFound)
		return
	}

	// Verify the request belongs to the user
	if request.UserID != userID {
		utils.SendErrorResponse(w, "Unauthorized to access this request", http.StatusForbidden)
		return
	}

	utils.SendSuccessResponse(w, "Financing request retrieved successfully", request, http.StatusOK)
}

// GetLatestFinancingRequest retrieves the latest financing request for the authenticated user
func (h *FinancingHandler) GetLatestFinancingRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.SendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := h.getUserIDFromRequest(r)
	if err != nil || userID == uuid.Nil {
		utils.SendErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	request, err := models.GetLatestFinancingRequestByUserID(h.DB, userID)
	if err != nil {
		utils.SendErrorResponse(w, "Database error", http.StatusInternalServerError)
		return
	}

	if request == nil {
		utils.SendSuccessResponse(w, "No financing request found", nil, http.StatusOK)
		return
	}

	utils.SendSuccessResponse(w, "Latest financing request retrieved successfully", request, http.StatusOK)
}
