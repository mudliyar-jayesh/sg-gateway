package main

import (
	"log"
	"net/http"
	"os"
	middleware "sg-gateway/internal/middlewares"
	"sg-gateway/internal/proxy"
	utils "sg-gateway/pkg/util/config"
	/*
			keystore "sg-gateway/pkg/util/keystore"
		    "encoding/base64"
	*/)

func main() {
	configPath := os.Getenv("SG_GATEWAY_CONFIG")
	if configPath == "" {
		log.Fatal("Environment variable 'SG_GATEWAY_CONFIG' is not set.")
	}

	// Load configuration
	config, err := utils.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
	proxy.LoadServiceMappings(config.Gateway.Services)

	/*

		// Load the secret key and IV from the key file
		keyStore, err := keystore.LoadKeyStore(config.Gateway.KeyFile)
		if err != nil {
			log.Fatalf("Error loading key store: %v", err)
		}

		// Decode the base64-encoded key and IV
		secretKey, err := base64.StdEncoding.DecodeString(keyStore.SecretKey)
		if err != nil {
			log.Fatalf("Error decoding secret key: %v", err)
		}

		iv, err := base64.StdEncoding.DecodeString(keyStore.IV)
		if err != nil {
			log.Fatalf("Error decoding IV: %v", err)
		}

	*/
	// Initialize mux
	mux := http.NewServeMux()
	mux.Handle("/", middleware.Logging(http.HandlerFunc(proxy.ResolveUrl)))

	// Middleware for token validation and encryption/decryption
	/*	mux.Handle("/", middleware.TokenValidationMiddleware(
			proxy.ReverseProxy(config),
			config, secretKey, iv,
		))
	*/
	corsHandlers := middleware.CORSMiddleware(mux)

	log.Println("Gateway running on port 45080...")
	httpErr := http.ListenAndServe(":45080", corsHandlers)
	if httpErr != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}
