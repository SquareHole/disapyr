package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/squarehole/disapyr/internal"
)

func main() {
	app := fiber.New()

	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// Load HTML templates
	captureSecretHTML, err := os.ReadFile("cmd/ui/capture_secret.html")
	if err != nil {
		log.Fatalf("Error reading capture_secret.html: %v", err)
	}

	// Load BASE_URL from environment variables
	baseURL := os.Getenv("BASE_URL")
	uiHostPort := os.Getenv("UI_HOST_PORT")

	// Get the access token from the external API
	log.Println("Getting access token...")
	accessToken, err := internal.GetAccessToken()
	if err != nil {
		fmt.Println("Error getting access token:", err)
		return
	}

	// GET handler to serve the main page for capturing the secret.
	app.Get("/", func(c *fiber.Ctx) error {
		log.Printf("Serving capture_secret.html")
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(string(captureSecretHTML))
	})

	// POST handler to capture the secret and generate the one-time link.
	app.Post("/", func(c *fiber.Ctx) error {
		log.Printf("POST /")
		secret := c.FormValue("secret")

		if secret != "" {
			log.Println("Secret provided")
			apiURL := fmt.Sprintf("%s/secret", baseURL)
			log.Printf("API URL: %s\n", apiURL)
			jsonData := []byte(fmt.Sprintf(`{"secret":"%s"}`, secret))

			var tr *http.Transport
			if os.Getenv("GO_ENV") != "production" {
				tr = &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // Bypass TLS verification
				}
			} else {
				tr = &http.Transport{}
			}

			log.Print("Creating new HTTP client")
			client := &http.Client{Transport: tr}
			req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
			if err != nil {
				log.Printf("Error creating new request: %v", err)
				return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

			log.Printf("Making API call")
			resp, err := client.Do(req)
			if err != nil {
				log.Printf("Error during API call: %v", err)
				return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
			}
			defer resp.Body.Close()

			var apiResponse struct {
				Key string `json:"key"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
				log.Printf("Error decoding API response: %v", err)
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
		var tr *http.Transport
		if os.Getenv("GO_ENV") != "production" {
			tr = &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // Bypass TLS verification
			}
		} else {
			tr = &http.Transport{}
		}

		client := &http.Client{Transport: tr}

		req, err := http.NewRequest("GET", apiURL, nil)
		if err != nil {
			log.Printf("Error creating new request: %v", err)
			return displaySecretPage(c, "Error retrieving secret")
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Error during API GET call: %v", err)
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
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}

		return displaySecretPage(c, apiResponse.Secret)
	})

	fmt.Printf("Starting server on :%s", uiHostPort)
	log.Fatal(app.Listen(fmt.Sprintf(":%s", uiHostPort)))
}

// displaySecretPage renders an HTML page with a read-only textarea containing the provided content.
func displaySecretPage(c *fiber.Ctx, content string) error {

	displaySecretHTML, err := os.ReadFile("cmd/ui/display_secret.html")
	if err != nil {
		log.Fatalf("Error reading display_secret.html: %v", err)
	}
	html := fmt.Sprintf(string(displaySecretHTML), content)
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.SendString(html)
}
