package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/squarehole/disapyr/internal"
)

func getCertKeyPaths() (string, string, error) {
	httpsEnabled := os.Getenv("HTTPS_ENABLED") != "false"
	certPath := os.Getenv("CERT_PATH")
	keyPath := os.Getenv("KEY_PATH")

	if httpsEnabled {
		log.Println("HTTPS enabled")
		if certPath == "" || keyPath == "" {
			return "", "", fmt.Errorf("CERT_PATH and KEY_PATH must be set when HTTPS_ENABLED is true")
		}

		if _, err := os.Stat(certPath); os.IsNotExist(err) {
			return "", "", fmt.Errorf("CERT_PATH file does not exist: %s", certPath)
		}

		if _, err := os.Stat(keyPath); os.IsNotExist(err) {
			return "", "", fmt.Errorf("KEY_PATH file does not exist: %s", keyPath)
		}
	}

	return certPath, keyPath, nil
}

func main() {
	// Load environment variables from .env file.
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Retrieve rate limit from environment variable.
	rateLimitStr := os.Getenv("RATE_LIMIT")
	rateLimit := 10 // Default to 10 requests per second.
	if rateLimitStr != "" {
		var err error
		rateLimit, err = strconv.Atoi(rateLimitStr)
		if err != nil {
			log.Fatal("Invalid RATE_LIMIT value:", err)
		}
	}

	// Create a new Fiber app.
	app := fiber.New()

	// Initialize the database connection.
	db, err := internal.NewDatabaseConnection()
	if err != nil {
		log.Fatal(err)
	}

	// Create the rate limiter.
	limiter := internal.CreateRateLimiter(rateLimit)

	// Register the routes.
	internal.RegisterRoutes(app, db, limiter, os.Getenv("ENC_KEY"), 32) // Assuming keyLen is 32

	// Start the Fiber app.
	port := ":8080"
	certPath, keyPath, err := getCertKeyPaths()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Starting API server on port %s...\n", port)
	if os.Getenv("HTTPS_ENABLED") != "false" {
		log.Fatal(app.ListenTLS(port, certPath, keyPath))
	} else {
		log.Fatal(app.Listen(port))
	}
}
