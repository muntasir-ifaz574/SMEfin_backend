package utils

import (
	"encoding/json"
	"mime"
	"net/http"
	"strings"
)

// ParseFormData parses form-data or JSON from request
// Supports both multipart/form-data and application/json
func ParseFormData(r *http.Request, v interface{}) error {
	contentType := r.Header.Get("Content-Type")
	
	// Handle multipart/form-data
	if strings.HasPrefix(contentType, "multipart/form-data") {
		if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB max
			return err
		}
		
		// Use reflection or manual mapping based on struct type
		// For now, we'll handle it manually in each handler
		return nil
	}
	
	// Handle application/x-www-form-urlencoded
	if strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
		if err := r.ParseForm(); err != nil {
			return err
		}
		return nil
	}
	
	// Handle JSON (fallback)
	if strings.HasPrefix(contentType, "application/json") {
		return json.NewDecoder(r.Body).Decode(v)
	}
	
	// Try to parse as form-data anyway
	if err := r.ParseMultipartForm(32 << 20); err == nil {
		return nil
	}
	
	// Try URL-encoded form
	if err := r.ParseForm(); err == nil {
		return nil
	}
	
	// Default to JSON
	return json.NewDecoder(r.Body).Decode(v)
}

// GetFormValue gets a value from form data (multipart or url-encoded) or JSON
func GetFormValue(r *http.Request, key string) string {
	contentType, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	
	// Try multipart form
	if strings.HasPrefix(contentType, "multipart/form-data") {
		if r.MultipartForm != nil {
			if values := r.MultipartForm.Value[key]; len(values) > 0 {
				return values[0]
			}
		}
	}
	
	// Try URL-encoded form
	if strings.HasPrefix(contentType, "application/x-www-form-urlencoded") || r.Form != nil {
		if values := r.Form[key]; len(values) > 0 {
			return values[0]
		}
		if values := r.PostForm[key]; len(values) > 0 {
			return values[0]
		}
	}
	
	// Try query params as fallback
	if values := r.URL.Query()[key]; len(values) > 0 {
		return values[0]
	}
	
	return ""
}

// ParseFormDataToStruct parses form data into a struct
// This is a helper that works with form field names matching struct field names (lowercase)
func ParseFormDataToStruct(r *http.Request, v interface{}) error {
	contentType := r.Header.Get("Content-Type")
	
	// Handle JSON first
	if strings.HasPrefix(contentType, "application/json") {
		return json.NewDecoder(r.Body).Decode(v)
	}
	
	// Handle form data
	if strings.HasPrefix(contentType, "multipart/form-data") {
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			return err
		}
		form := r.MultipartForm.Value
		
		// Convert form data to JSON-like structure and decode
		formMap := make(map[string]interface{})
		for key, values := range form {
			if len(values) > 0 {
				formMap[key] = values[0]
			}
		}
		
		jsonData, err := json.Marshal(formMap)
		if err != nil {
			return err
		}
		
		return json.Unmarshal(jsonData, v)
	}
	
	// Handle URL-encoded form
	if strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
		if err := r.ParseForm(); err != nil {
			return err
		}
		
		formMap := make(map[string]interface{})
		for key, values := range r.PostForm {
			if len(values) > 0 {
				formMap[key] = values[0]
			}
		}
		
		jsonData, err := json.Marshal(formMap)
		if err != nil {
			return err
		}
		
		return json.Unmarshal(jsonData, v)
	}
	
	// Default to JSON
	return json.NewDecoder(r.Body).Decode(v)
}

