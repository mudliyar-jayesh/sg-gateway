package proxy

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

var services map[string]string

func LoadServiceMappings(configServices map[string]string) {
	services = configServices
}

func proxyRequest(method, destination string, r *http.Request) (string, error) {
	if r.URL.RawQuery != "" {
		destination = destination + "?" + r.URL.RawQuery
	}

	request, err := http.NewRequest(method, destination, r.Body)
	if err != nil {
		return "", err
	}

	request.Header = r.Header

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return string(bodyBytes), nil
}

func ResolveUrl(w http.ResponseWriter, r *http.Request) {
	/*services := map[string]string{
		"/api/portal": "http://localhost:8080",
	}*/

	var wg sync.WaitGroup

	responseSent := make(chan bool) // Channel to notify when a response has been sent
	var responseOnce sync.Once      // To ensure response is sent only once

	// Forward the request to all matching services
	for path, target := range services {
		if !strings.Contains(r.URL.Path, path) {
			continue
		}
		newPath := strings.TrimPrefix(r.URL.Path, path)
		targetUrl := fmt.Sprintf("%s%s", target, newPath)

		wg.Add(1)

		go func(targetUrl string) {
			defer wg.Done()

			// Forward the request to the target service
			log.Printf("[==>] Forwarding URL: %s\n", targetUrl)
			responseBody, err := proxyRequest(r.Method, targetUrl, r)
			if err != nil {
				log.Printf("[!] Error from %s: %v\n", targetUrl, err)
				return
			}

			// Ensure only one response is sent to the client
			responseOnce.Do(func() {
				log.Printf("[<==] Forwarding response from %s\n", targetUrl)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(responseBody))
				responseSent <- true // Notify that a response was sent
			})
		}(targetUrl)
	}

	// Wait for all Goroutines to finish
	go func() {
		wg.Wait()
		close(responseSent) // Close the channel after all requests are processed
	}()

	// Use select to handle response or timeout
	select {
	case <-responseSent: // If a response was successfully sent
		log.Printf("[-] Response sent, exiting...")
	case <-time.After(10 * time.Second): // Timeout case
		http.Error(w, "No response from service", http.StatusGatewayTimeout)
	}
}
