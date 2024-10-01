package proxy

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	utils "sg-gateway/pkg/util/config"
	"strings"
)

func ReverseProxy(config *utils.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var baseURL string
		var newPath string

		// Determine the base URL of the target backend service based on the URL path
		if strings.HasPrefix(r.URL.Path, "/api/portal") {
			baseURL = config.Gateway.Services["sg_portal"]
			// Remove the "/api/portal" prefix from the URL path
			newPath = strings.TrimPrefix(r.URL.Path, "/api/portal")
		} else if strings.HasPrefix(r.URL.Path, "/api/bmrm") {
			baseURL = config.Gateway.Services["bmrm"]
			// Remove the "/api/bmrm" prefix from the URL path
			newPath = strings.TrimPrefix(r.URL.Path, "/api/bmrm")
		} else {
			// If no matching service is found, return 404
			http.Error(w, "Service not found", http.StatusNotFound)
			log.Println("Service not found for URL:", r.URL.Path)
			return
		}

		// Parse the base URL (backend service URL)
		backendURL, err := url.Parse(baseURL)
		if err != nil {
			http.Error(w, "Invalid base URL for backend service", http.StatusInternalServerError)
			log.Printf("Error parsing base URL: %v", err)
			return
		}

		// Log the backend base URL
		log.Printf("[Backend Base URL]: %s", backendURL.String())

		// Create a reverse proxy that forwards the request to the backend URL
		proxy := httputil.NewSingleHostReverseProxy(backendURL)

		// Customize the Director function
		originalDirector := proxy.Director
		proxy.Director = func(req *http.Request) {
			originalDirector(req)
			// Set the correct path and query
			req.URL.Path = singleJoiningSlash(backendURL.Path, newPath)
			req.URL.RawQuery = r.URL.RawQuery
			// Set the Host header to the backend host
			req.Host = backendURL.Host
			// Copy the headers from the original request
			req.Header = r.Header.Clone()
		}

		// Modify the response to log the status code from the backend
		proxy.ModifyResponse = func(response *http.Response) error {
			log.Printf("[Response from Backend] Status: %d | message: %v", response.StatusCode, response.Body)
			return nil
		}

		// Forward the original request to the backend service
		proxy.ServeHTTP(w, r)
	}
}

// Helper function to join paths
func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}
