package proxy

import (
	"io"
	"log"
	"net/http"
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
	request.ContentLength = r.ContentLength
	request.TransferEncoding = r.TransferEncoding

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
	var targetUrl = r.Header.Get("targetUrl")
	log.Printf("[==>] Forwarding URL: %s\n", targetUrl)
	responseBody, err := proxyRequest(r.Method, targetUrl, r)
	if err != nil {
		log.Printf("[!] Error from %s: %v\n", targetUrl, err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	// Send the response back to the client
	log.Printf("[<==] Forwarding response from %s\n", targetUrl)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(responseBody))
}
