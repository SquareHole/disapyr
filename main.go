package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/squarehole/disapyr/internal"
	"golang.org/x/time/rate"
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

	// Create a rate limiter.
	Limiter := createRateLimiter(rateLimit)

	httpsEnabledStr := os.Getenv("HTTPS_ENABLED")
	httpsEnabled := httpsEnabledStr != "false"

	// Retrieve database connection details from environment variables.
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	encKey := os.Getenv("ENC_KEY")
	keyLenStr := os.Getenv("KEY_LEN")
	keyLen, err := strconv.Atoi(keyLenStr)
	if err != nil {
		log.Fatal("Invalid KEY_LEN value:", err)
	}

	// Build the PostgreSQL connection string.
	var connStr string
	if dbPassword != "" {
		connStr = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbPort, dbName)
	} else {
		connStr = fmt.Sprintf("postgres://%s@%s:%s/%s?sslmode=disable", dbUser, dbHost, dbPort, dbName)
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}
	defer db.Close()

	// Create the 'secrets' table if it does not exist.
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS secrets (
		key TEXT PRIMARY KEY,
		secret TEXT,
		retrieved_at TIMESTAMP NULL
	);
	`
	if _, err := db.Exec(createTableSQL); err != nil {
		log.Fatal("Error creating table:", err)
	}

	// Create a new Fiber app.
	app := fiber.New()

	// Endpoint to store a secret.
	app.Post("/secret", func(c *fiber.Ctx) error {
		// Limit the number of requests.
		if !Limiter.Allow() {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"error": "too many requests"})
		}

		type RequestBody struct {
			Secret string `json:"secret"`
		}
		var body RequestBody
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON"})
		}
		if body.Secret == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "secret is required"})
		}

		// Generate a unique key.
		uid := uuid.New().String()
		key, err := internal.HideIdentifier(uid, []byte(encKey))
		if len(key) > keyLen {
			key = key[:keyLen]
		}
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate key"})
		}

		// Insert the secret and key into the database.
		_, err = db.Exec("INSERT INTO secrets(key, secret) VALUES($1, $2)", key, body.Secret)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to store secret"})
		}

		return c.JSON(fiber.Map{"key": key})
	})

	// Endpoint to retrieve a secret exactly once.
	app.Get("/secret/:key", func(c *fiber.Ctx) error {
		// Limit the number of requests.
		if !Limiter.Allow() {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"error": "too many requests"})
		}

		key := c.Params("key")

		// Start a transaction to ensure atomic read-update.
		tx, err := db.Begin()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to start transaction"})
		}
		defer tx.Rollback()

		var secret string
		var retrievedAt *time.Time
		err = tx.QueryRow("SELECT secret, retrieved_at FROM secrets WHERE key = $1 FOR UPDATE", key).Scan(&secret, &retrievedAt)
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "secret not found"})
		} else if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to query secret"})
		}

		// Check if the secret has already been retrieved.
		if retrievedAt != nil || secret == "" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "secret already retrieved"})
		}

		// Clear the secret and record the retrieval time.
		now := time.Now()
		_, err = tx.Exec("UPDATE secrets SET secret = '', retrieved_at = $1 WHERE key = $2", now, key)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update secret"})
		}

		// Commit the transaction.
		if err = tx.Commit(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to commit transaction"})
		}

		// Return the original secret.
		return c.JSON(fiber.Map{"secret": secret})
	})

	// Start the Fiber app.
	port := ":3000"
	if httpsEnabled {
		log.Fatal(app.ListenTLS(port, "cert.pem", "key.pem"))
	} else {
		log.Fatal(app.Listen(port))
	}
}

func createRateLimiter(rateLimit int) *rate.Limiter {
	return rate.NewLimiter(rate.Limit(rateLimit), 2*rateLimit)
}
