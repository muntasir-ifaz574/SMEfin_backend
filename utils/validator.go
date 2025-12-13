package utils

import (
	"mime"
	"path/filepath"
	"regexp"
	"strings"
)

func ValidateEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func ValidatePhone(phone string) bool {
	// Remove common formatting characters
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	phone = strings.ReplaceAll(phone, "(", "")
	phone = strings.ReplaceAll(phone, ")", "")
	
	// Check if it's a valid phone number (at least 10 digits)
	phoneRegex := regexp.MustCompile(`^\d{10,15}$`)
	return phoneRegex.MatchString(phone)
}

func ValidateOTP(otp string) bool {
	otpRegex := regexp.MustCompile(`^\d{6}$`)
	return otpRegex.MatchString(otp)
}

// ValidateFileType checks if the file extension is allowed
func ValidateFileType(filename string, allowedTypes []string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return false
	}
	// Remove the dot
	ext = ext[1:]
	
	for _, allowedType := range allowedTypes {
		if strings.ToLower(allowedType) == ext {
			return true
		}
	}
	return false
}

// ValidateFileSize checks if file size is within limit (size in bytes)
func ValidateFileSize(fileSize int64, maxSizeMB int) bool {
	maxSizeBytes := int64(maxSizeMB) * 1024 * 1024
	return fileSize <= maxSizeBytes && fileSize > 0
}

// GetFileMimeType returns the MIME type based on file extension
func GetFileMimeType(filename string) string {
	ext := filepath.Ext(filename)
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		return "application/octet-stream"
	}
	return mimeType
}

