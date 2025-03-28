package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/charmbracelet/log"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/squarehole/disapyr/internal"
)

func main() {

	logger := log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    true,
		ReportTimestamp: true,
		TimeFormat:      time.RFC3339Nano,
		Level:           log.DebugLevel,
		Formatter:       log.JSONFormatter,
	})

	log.SetDefault(logger)

	app := fiber.New()

	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	root, err := os.OpenRoot("./")
	if err != nil {
		log.Fatal("Error opening root:", "err", err)
	}

	defer root.Close()

	// Serve static files
	app.Static(".", root.Name())

	captureSecretHTML, err := os.ReadFile("capture_secret.html")
	if err != nil {
		log.Fatal("Error reading capture_secret.html: %v", err)
	}

	// Load BASE_URL from environment variables
	baseURL := os.Getenv("BASE_URL")
	uiHostPort := os.Getenv("UI_HOST_PORT")

	// Get the access token from the external API
	log.Info("Getting access token...")
	accessToken, err := internal.GetAccessToken()
	if err != nil {
		log.Error("Error getting access token:", "err", err)
		return
	}

	// GET handler to serve the main page for capturing the secret.
	app.Get("/", func(c *fiber.Ctx) error {
		log.Info("Serving capture_secret.html")
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(string(captureSecretHTML))
	})

	// POST handler to capture the secret and generate the one-time link.
	app.Post("/", func(c *fiber.Ctx) error {
		log.Info("POST /")
		secret := c.FormValue("secret")

		if secret != "" {
			log.Info("Secret provided")
			apiURL := fmt.Sprintf("%s/secret", baseURL)
			log.Info("API URL: ", "url", apiURL)
			jsonData := []byte(fmt.Sprintf(`{"secret":"%s"}`, secret))

			// Create HTTP client with secure TLS configuration
			client := createSecureHTTPClient()

			log.Info("Creating new HTTP client")
			req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
			if err != nil {
				log.Error("Error creating request", "err", err)
				return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

			log.Info("Making API call")
			resp, err := client.Do(req)
			if err != nil {
				log.Error("Error during API call", "err", err)
				return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
			}
			defer resp.Body.Close()

			var apiResponse struct {
				Key string `json:"key"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
				log.Error("Error decoding API response", "err", err)
				return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
			}

			// Get the current URL for the app hosted by Fiber
			currentURL := fmt.Sprintf("%s://%s/secret/%s", c.Protocol(), c.Hostname(), apiResponse.Key)
			fmt.Printf("Current URL: %s\n", currentURL)

			// The external API returns a key which is used to build the one-time link.
			c.Set("Content-Type", "text/html; charset=utf-8")
			return c.SendString(fmt.Sprintf("<div>%s</div>", currentURL))
		}
		return c.SendString("No secret provided")
	})

	// GET handler for displaying a secret retrieved from the external API.
	app.Get("/secret/:key", func(c *fiber.Ctx) error {
		key := c.Params("key")
		apiURL := fmt.Sprintf("%s/secret/%s", baseURL, key)

		// Create HTTP client with secure TLS configuration
		client := createSecureHTTPClient()

		req, err := http.NewRequest("GET", apiURL, nil)
		if err != nil {
			log.Error("Error creating request", "err", err)
			return displaySecretPage(c, "Error retrieving secret")
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

		resp, err := client.Do(req)
		if err != nil {
			log.Error("Error during API call", "err", err)
			return displaySecretPage(c, "Error retrieving secret")
		}
		defer resp.Body.Close()

		// If the API returns a non-200 status, display a message in the readonly textbox.
		if resp.StatusCode != http.StatusOK {
			message := "Secret not found. It may have already been retrieved."
			return displaySecretPage(c, message)
		}

		var apiResponse struct {
			Secret string `json:"secret"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
			log.Error("Error decoding API response", "err", err)
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}

		return displaySecretPage(c, apiResponse.Secret)
	})

	log.Info("Starting server on:", "port", uiHostPort)
	log.Fatal(app.Listen(fmt.Sprintf(":%s", uiHostPort)))
}

// displaySecretPage renders an HTML page with a read-only textarea containing the provided content.
func displaySecretPage(c *fiber.Ctx, content string) error {
	displaySecretHTML, err := os.ReadFile("display_secret.html")
	if err != nil {
		log.Error("Error reading display_secret.html", "err", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Error loading display template")
	}
	html := fmt.Sprintf(string(displaySecretHTML), content)
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.SendString(html)
}

// createSecureHTTPClient creates an HTTP client with secure TLS configuration
func createSecureHTTPClient() *http.Client {
	var tr *http.Transport

	// Check if we're in production mode
	if os.Getenv("GO_ENV") == "production" {
		// In production, use the default transport with standard TLS verification
		tr = &http.Transport{}
		log.Info("Using production TLS configuration")
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
				log.Warn("Could not read custom CA certificate", "err", err)
				log.Info("Falling back to standard TLS verification")
				tr = &http.Transport{}
			} else {
				// Add the custom CA certificate to the cert pool
				if ok := rootCAs.AppendCertsFromPEM(caCert); !ok {
					log.Warn("Failed to append custom CA certificate to cert pool")
					log.Info("Falling back to standard TLS verification")
					tr = &http.Transport{}
				} else {
					// Use the custom CA certificate for TLS verification
					tr = &http.Transport{
						TLSClientConfig: &tls.Config{
							RootCAs: rootCAs,
						},
					}
					log.Info("Using custom CA certificate for TLS verification")
				}
			}
		} else {
			// No custom CA certificate provided, use standard TLS verification
			// but log a warning
			log.Warn("No custom CA certificate provided for development environment")
			log.Info("Using standard TLS verification. Set CUSTOM_CA_CERT environment variable to use a custom CA certificate")
			tr = &http.Transport{}
		}
	}

	return &http.Client{Transport: tr}
}
