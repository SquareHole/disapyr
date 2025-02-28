package main

import (
	"log"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/squarehole/disapyr/internal"
)

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
	port := ":3000"
	if os.Getenv("HTTPS_ENABLED") != "false" {
		log.Fatal(app.ListenTLS(port, "cert.pem", "key.pem"))
	} else {
		log.Fatal(app.Listen(port))
	}
}
