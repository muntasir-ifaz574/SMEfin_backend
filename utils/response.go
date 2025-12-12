package utils

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Success    bool        `json:"success"`
	Message    string      `json:"message"`
	StatusCode int         `json:"status_code"`
	Data       interface{} `json:"data,omitempty"`
}

func SendSuccessResponse(w http.ResponseWriter, message string, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	response := Response{
		Success:    true,
		Message:    message,
		StatusCode: statusCode,
		Data:       data,
	}
	
	json.NewEncoder(w).Encode(response)
}

func SendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	response := Response{
		Success:    false,
		Message:    message,
		StatusCode: statusCode,
	}
	
	json.NewEncoder(w).Encode(response)
}

