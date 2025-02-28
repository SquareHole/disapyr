package internal

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/mr-tron/base58"
	"golang.org/x/time/rate"
)

// HideIdentifier encrypts the provided identifier using AES-GCM,
// prepends the nonce, and returns a Base58-encoded string.
func HideIdentifier(id string, key []byte) (string, error) {
	// AES-GCM requires a nonce. The recommended size is 12 bytes.
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Create a new AES cipher using the key.
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create the AES-GCM cipher mode instance.
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Encrypt the identifier.
	encrypted := aead.Seal(nil, nonce, []byte(id), nil)

	// Concatenate the nonce and the encrypted data.
	combined := append(nonce, encrypted...)

	encoded := base58.Encode(combined)
	return encoded, nil
}

func NewDatabaseConnection() (*sql.DB, error) {
	// Retrieve database connection details from environment variables.
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	// Build the PostgreSQL connection string.
	var connStr string
	if dbPassword != "" {
		connStr = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbPort, dbName)
	} else {
		connStr = fmt.Sprintf("postgres://%s@%s:%s/%s?sslmode=disable", dbUser, dbHost, dbPort, dbName)
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	// Create the 'secrets' table if it does not exist.
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS secrets (
		key TEXT PRIMARY KEY,
		secret TEXT,
		retrieved_at TIMESTAMP NULL
	);
	`
	if _, err := db.Exec(createTableSQL); err != nil {
		return nil, fmt.Errorf("error creating table: %w", err)
	}

	return db, nil
}

func CreateRateLimiter(rateLimit int) *rate.Limiter {
	return rate.NewLimiter(rate.Limit(rateLimit), 2*rateLimit)
}

func RegisterRoutes(app *fiber.App, db *sql.DB, limiter *rate.Limiter, encKey string, keyLen int) {
	// Endpoint to store a secret.
	app.Post("/secret", func(c *fiber.Ctx) error {
		// Limit the number of requests.
		if !limiter.Allow() {
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
		key, err := HideIdentifier(uid, []byte(encKey))
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
		if !limiter.Allow() {
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
}
