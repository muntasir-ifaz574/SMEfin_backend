package storage

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// UploadFileToSupabase uploads a file to Supabase storage bucket
func UploadFileToSupabase(file multipart.File, filename string, bucketName string) (string, error) {
	// Get Supabase credentials from environment
	supabaseURL := os.Getenv("SUPABASE_URL")
	// Try service role key first (for server-side uploads), fallback to anon key
	supabaseKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	if supabaseKey == "" {
		supabaseKey = os.Getenv("SUPABASE_ANON_KEY")
	}

	if supabaseURL == "" {
		return "", fmt.Errorf("SUPABASE_URL environment variable is required")
	}
	if supabaseKey == "" {
		return "", fmt.Errorf("SUPABASE_SERVICE_ROLE_KEY or SUPABASE_ANON_KEY environment variable is required")
	}

	// Generate unique filename
	ext := filepath.Ext(filename)
	uniqueFilename := fmt.Sprintf("%s_%d%s", uuid.New().String(), time.Now().Unix(), ext)

	// Read file content
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Create the upload URL
	uploadURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", supabaseURL, bucketName, uniqueFilename)

	// Create HTTP request
	req, err := http.NewRequest("POST", uploadURL, bytes.NewReader(fileBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+supabaseKey)
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("x-upsert", "true") // Allow overwriting

	// Make the request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Return the public URL
	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", supabaseURL, bucketName, uniqueFilename)
	return publicURL, nil
}

// GetSupabasePublicURL generates the public URL for a file in Supabase storage
func GetSupabasePublicURL(filename string, bucketName string) string {
	supabaseURL := os.Getenv("SUPABASE_URL")
	if supabaseURL == "" {
		return ""
	}
	return fmt.Sprintf("%s/storage/v1/object/public/%s/%s", supabaseURL, bucketName, filename)
}
