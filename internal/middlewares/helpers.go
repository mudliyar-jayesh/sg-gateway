package middleware

import (
	"bytes"
	"encoding/json"
	//	"fmt"
	"io"
	"log"
	"net/http"
	utils "sg-gateway/pkg/util/config"
	"sg-gateway/pkg/util/encryption"
	"strings"
	"time"
)

type ValidationResponse struct {
	Valid  bool   `json:"valid"`
	UserID string `json:"user_id"`
}

// Middleware to handle CORS
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // Allow all origins
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Priority, companyid, token")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// Struct to handle the expected incoming request format
type EncryptedRequest struct {
	Data string `json:"data"`
}

func TokenValidationMiddleware(next http.Handler, config *utils.Config, secretKey, iv []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log the incoming URL path
		log.Printf("\n[INFO]::['URL: %v']", r.URL.Path)

		var bodyBytes []byte
		var err error

		// Read the original request body
		if r.Body != nil {
			bodyBytes, err = io.ReadAll(r.Body)
			if err != nil {
				log.Printf("\n[ERR]::['Invalid Request Body: %v']", err)
				http.Error(w, "Failed to read request body", http.StatusInternalServerError)
				return
			}
			// Close the original body to prevent resource leaks
			r.Body.Close()
		}

		// Decrypt the request body (applies to all paths)
		if len(bodyBytes) > 0 {
			// Parse the JSON to extract the `data` field
			var encryptedReq struct {
				Data string `json:"data"`
			}
			err = json.Unmarshal(bodyBytes, &encryptedReq)
			if err != nil {
				log.Printf("\n[ERR]::['Invalid JSON Format: %v']", err)
				http.Error(w, "Invalid JSON format", http.StatusBadRequest)
				return
			}

			// Decrypt the `data` field
			decryptedData, err := encryption.Decrypt(encryptedReq.Data, secretKey, iv)
			if err != nil {
				log.Printf("\n[ERR]::['Decryption Failed: %v']", err)
				http.Error(w, "Failed to decrypt request body", http.StatusInternalServerError)
				return
			}

			// Debugging: Log the decrypted data
			log.Printf("\n[INFO]::['Decrypted Data']: %s", decryptedData)

			// Check if the decrypted data is valid JSON
			if !json.Valid([]byte(decryptedData)) {
				log.Printf("\n[ERR]::['Decrypted data is not valid JSON']")
				http.Error(w, "Decrypted data is not valid JSON", http.StatusInternalServerError)
				return
			}

			// Replace the request body with the decrypted data
			r.Body = io.NopCloser(strings.NewReader(decryptedData))

			// Update the Content-Length header
			r.ContentLength = int64(len(decryptedData))

			// Optionally, update the Content-Type header if necessary
			r.Header.Set("Content-Type", "application/json")

			// Set the GetBody function so that ReverseProxy can re-read the body if needed
			r.GetBody = func() (io.ReadCloser, error) {
				return io.NopCloser(strings.NewReader(decryptedData)), nil
			}
		} else {
			// If there was no body, reset it to an empty reader
			r.Body = io.NopCloser(bytes.NewReader(nil))

			// Set the GetBody function to return an empty reader
			r.GetBody = func() (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader(nil)), nil
			}
		}

		// Now handle token validation, but only for non-excluded paths
		for _, path := range config.Gateway.ExcludedPaths {
			if strings.HasPrefix(r.URL.Path, path) {
				// If the path is excluded, no token validation
				next.ServeHTTP(w, r)
				return
			}
		}

		// Check for token and company headers for non-excluded paths
		token := r.Header.Get("token")
		company := r.Header.Get("companyid")
		if token == "" || company == "" {
			log.Printf("\n[ERR]::['Missing Headers']")
			http.Error(w, "Missing required headers", http.StatusUnauthorized)
			return
		}

		// Validate the token
		validationRes, err := validateToken(token, config.Gateway.SgPortalURL)
		if err != nil || !validationRes.Valid {
			log.Printf("\n[ERR]::['Invalid Token: %v']", err)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Remove token header and add UserID header
		r.Header.Del("token")
		r.Header.Set("userid", validationRes.UserID)

		log.Printf("\n[INFO]::['Token Validated, Requesting internal service']")
		next.ServeHTTP(w, r)
	})
}

// Function to validate token by calling SG Portal Service
func validateToken(token, sgPortalURL string) (*ValidationResponse, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", sgPortalURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("token", token)
	res, err := client.Do(req)
	if err != nil {
		log.Printf("\n[ERR]::['Token Validation Failed']")
		return nil, err
	}
	defer res.Body.Close()

	var validationRes ValidationResponse
	if err := json.NewDecoder(res.Body).Decode(&validationRes); err != nil {
		log.Printf("\n[ERR]::['Invalid Token Validation Response']")
		return nil, err
	}

	return &validationRes, nil
}
