// Package main provides a CLI tool for securely storing and retrieving secrets
// via an API server. The tool supports two main operations: storing a secret
// and retrieving a secret using a key.
//
// Usage:
//   - To store a secret:
//     ./cli --store --secret="your-secret" --server="https://your-server-url"
//   - To retrieve a secret:
//     ./cli --retrieve --key="your-key" --server="https://your-server-url"
//
// Flags:
//
//	--store       Indicates that a secret should be stored.
//	--retrieve    Indicates that a secret should be retrieved.
//	--secret      The secret to store (used with --store).
//	--key         The key to retrieve the secret (used with --retrieve).
//	--server      The API server URL (default: https://localhost:3000).
//
// Environment Variables:
//
//	GO_ENV           If set to "production", TLS verification will be enforced.
//	CUSTOM_CA_CERT   Path to a custom CA certificate for development environments.
//	                 If provided, this certificate will be used instead of bypassing TLS verification.
//
// API Endpoints:
//   - POST /secret: Stores a secret and returns a key.
//   - GET /secret/:key: Retrieves a secret using the provided key.
//
// Error Handling:
//   - The tool ensures that either --store or --retrieve is specified, but not both.
//   - Input validation is performed for required flags (--secret for storing and --key for retrieving).
//   - Errors from API calls or JSON processing are displayed, and the program exits with a non-zero status.
//
// Example:
//
//	To store a secret:
//	  ./cli --store --secret="my-secret" --server="https://api.example.com"
//	To retrieve a secret:
//	  ./cli --retrieve --key="my-key" --server="https://api.example.com"
package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
)

func main() {
	// Define CLI flags.
	storeCmd := flag.Bool("store", false, "Store a secret")
	retrieveCmd := flag.Bool("retrieve", false, "Retrieve a secret")
	secretVal := flag.String("secret", "", "Secret to store (use with --store)")
	keyVal := flag.String("key", "", "Key to retrieve the secret (use with --retrieve)")
	server := flag.String("server", "https://localhost:3000", "API server URL (default: https://localhost:3000)")
	flag.Parse()

	// Ensure that either store or retrieve is selected, but not both.
	if *storeCmd && *retrieveCmd {
		fmt.Println("Error: Cannot use both --store and --retrieve flags together.")
		os.Exit(1)
	}
	if !*storeCmd && !*retrieveCmd {
		fmt.Println("Error: Please specify either --store or --retrieve flag.")
		flag.Usage()
		os.Exit(1)
	}

	// Create HTTP client with appropriate TLS configuration
	client := createHTTPClient()

	if *storeCmd {
		// Validate input.
		if *secretVal == "" {
			fmt.Println("Error: Please provide a secret using the --secret flag.")
			os.Exit(1)
		}

		// Prepare and send a POST request to /secret.
		url := fmt.Sprintf("%s/secret", *server)
		payload := map[string]string{"secret": *secretVal}
		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			fmt.Println("Error marshalling JSON:", err)
			os.Exit(1)
		}

		resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
		if err != nil {
			fmt.Println("Error calling API:", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response:", err)
			os.Exit(1)
		}
		if resp.StatusCode != http.StatusOK {
			fmt.Printf("API error: %s\n", body)
			os.Exit(1)
		}

		// Display the returned key.
		fmt.Printf("Secret stored successfully. Key: %s\n", body)
	} else if *retrieveCmd {
		// Validate input.
		if *keyVal == "" {
			fmt.Println("Error: Please provide a key using the --key flag.")
			os.Exit(1)
		}

		// Prepare and send a GET request to /secret/:key.
		url := fmt.Sprintf("%s/secret/%s", *server, *keyVal)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Println("Error creating request:", err)
			os.Exit(1)
		}

		// Get access token for authentication
		// Note: This is a placeholder. In a real implementation, you would need to
		// obtain an access token from Auth0 or another authentication provider.
		// For now, we're just making an unauthenticated request.

		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error calling API:", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response:", err)
			os.Exit(1)
		}
		if resp.StatusCode != http.StatusOK {
			fmt.Printf("API error: %s\n", body)
			os.Exit(1)
		}

		// Display the retrieved secret.
		fmt.Printf("Retrieved secret: %s\n", body)
	}
}

// createHTTPClient creates an HTTP client with appropriate TLS configuration
func createHTTPClient() *http.Client {
	var tr *http.Transport

	// Check if we're in production mode
	if os.Getenv("GO_ENV") == "production" {
		// In production, use the default transport with standard TLS verification
		tr = &http.Transport{}
	} else {
		// In development, check if a custom CA certificate is provided
		customCACert := os.Getenv("CUSTOM_CA_CERT")
		if customCACert != "" {
			// Use the custom CA certificate
			rootCAs, _ := x509.SystemCertPool()
			if rootCAs == nil {
				rootCAs = x509.NewCertPool()
			}

			// Read the custom CA certificate
			caCert, err := os.ReadFile(customCACert)
			if err != nil {
				fmt.Printf("Warning: Could not read custom CA certificate: %v\n", err)
				fmt.Println("Falling back to standard TLS verification")
				tr = &http.Transport{}
			} else {
				// Add the custom CA certificate to the cert pool
				if ok := rootCAs.AppendCertsFromPEM(caCert); !ok {
					fmt.Println("Warning: Failed to append custom CA certificate to cert pool")
					fmt.Println("Falling back to standard TLS verification")
					tr = &http.Transport{}
				} else {
					// Use the custom CA certificate for TLS verification
					tr = &http.Transport{
						TLSClientConfig: &tls.Config{
							RootCAs: rootCAs,
						},
					}
					fmt.Println("Using custom CA certificate for TLS verification")
				}
			}
		} else {
			// No custom CA certificate provided, use standard TLS verification
			// but print a warning
			fmt.Println("Warning: No custom CA certificate provided for development environment")
			fmt.Println("Using standard TLS verification. Set CUSTOM_CA_CERT environment variable to use a custom CA certificate")
			tr = &http.Transport{}
		}
	}

	return &http.Client{Transport: tr}
}
