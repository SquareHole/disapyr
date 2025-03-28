package internal

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/log"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/mr-tron/base58"
	"golang.org/x/time/rate"
)

var handler = func(c *fiber.Ctx) error {
	log.Info("Handler called")

	// Get the token from the Authorization header.
	tokenString := c.Get("Authorization")
	if tokenString == "" {
		return HandleAuthError(c, "Missing authentication token", nil)
	}

	// Trim and remove "Bearer " prefix (case-insensitive) if present.
	tokenString = strings.TrimSpace(tokenString)
	if strings.HasPrefix(strings.ToLower(tokenString), "bearer ") {
		tokenString = tokenString[len("Bearer "):]
	}

	// Validate the token with Auth0.
	auth0Domain := os.Getenv("URL")
	if auth0Domain == "" {
		return HandleServerError(c, "Auth0 domain not configured", nil)
	}

	// Get the audience from environment variable
	audience := os.Getenv("AUDIENCE")
	if audience == "" {
		return HandleServerError(c, "Audience not configured", nil)
	}

	log.Infof("Auth0 domain: %s, Audience: %s", auth0Domain, audience)
	jwksURL := fmt.Sprintf("https://%s/.well-known/jwks.json", auth0Domain)

	// Parse with audience and issuer validation
	parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing method is RSA.
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			log.Warnf("unexpected signing method: %v", token.Header["alg"])
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Fetch the JWKS.
		resp, err := http.Get(jwksURL)
		if err != nil {
			log.Errorf("failed to fetch JWKS: %v", err)
			return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
		}
		defer resp.Body.Close()

		// Check if the status code is OK.
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			log.Errorf("failed to fetch JWKS, status %d: %s", resp.StatusCode, string(body))
			return nil, fmt.Errorf("failed to fetch JWKS, status %d", resp.StatusCode)
		}

		var jwks struct {
			Keys []json.RawMessage `json:"keys"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
			log.Errorf("failed to decode JWKS: %v", err)
			return nil, fmt.Errorf("failed to decode JWKS: %w", err)
		}

		// Look for a key that matches the token's "kid" header.
		for _, key := range jwks.Keys {
			var k struct {
				Kid string `json:"kid"`
				N   string `json:"n"`
				E   string `json:"e"`
			}
			if err := json.Unmarshal(key, &k); err != nil {
				continue
			}
			if k.Kid == token.Header["kid"] {
				nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
				if err != nil {
					log.Errorf("failed to decode N: %v", err)
					return nil, fmt.Errorf("failed to decode N: %w", err)
				}
				eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
				if err != nil {
					log.Errorf("failed to decode E: %v", err)
					return nil, fmt.Errorf("failed to decode E: %w", err)
				}
				e := 0
				for _, b := range eBytes {
					e = e*256 + int(b)
				}
				return &rsa.PublicKey{
					N: new(big.Int).SetBytes(nBytes),
					E: e,
				}, nil
			}
		}
		log.Warn("no matching key found")
		return nil, fmt.Errorf("no matching key found")
	})

	if err != nil {
		return HandleAuthError(c, "Invalid token", err)
	}

	// In jwt/v5 we need to explicitly check claims validity
	if !parsedToken.Valid {
		return HandleAuthError(c, "Token validation failed", nil)
	}

	// Validate audience and issuer claims
	claims := parsedToken.Claims.(jwt.MapClaims)

	// Log the claims for debugging
	log.Debugf("Token claims: %v", claims)

	// Verify audience claim
	tokenAud, ok := claims["aud"].(string)
	if !ok {
		// Handle case where aud is an array of strings
		tokenAudArray, ok := claims["aud"].([]interface{})
		if !ok {
			return HandleAuthError(c, fmt.Sprintf("Invalid audience format: %v", claims["aud"]), nil)
		}

		// Check if audience is in the array
		audFound := false
		for _, aud := range tokenAudArray {
			if audStr, ok := aud.(string); ok && audStr == audience {
				audFound = true
				break
			}
		}

		if !audFound {
			return HandleAuthError(c, fmt.Sprintf("Invalid audience: %v, expected: %s", tokenAudArray, audience), nil)
		}
	} else if tokenAud != audience {
		return HandleAuthError(c, fmt.Sprintf("Invalid audience: %v, expected: %s", tokenAud, audience), nil)
	}

	// Verify issuer claim
	expectedIssuer := fmt.Sprintf("https://%s/", auth0Domain)
	tokenIss, ok := claims["iss"].(string)
	if !ok || tokenIss != expectedIssuer {
		return HandleAuthError(c, fmt.Sprintf("Invalid issuer: %v, expected: %s", claims["iss"], expectedIssuer), nil)
	}

	// Token is valid; proceed to the next handler.
	return c.Next()
}

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
	dbUseSSL := os.Getenv("DB_USESSL")
	sslMode := "disable"
	if strings.ToLower(dbUseSSL) == "true" {
		sslMode = "require"
	}

	// Build the PostgreSQL connection string.
	var connStr string
	if dbPassword != "" {
		connStr = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", dbUser, dbPassword, dbHost, dbPort, dbName, sslMode)
	} else {
		connStr = fmt.Sprintf("postgres://%s@%s:%s/%s?sslmode=%s", dbUser, dbHost, dbPort, dbName, sslMode)
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	// Configure connection pooling and timeouts
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

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

	log.Info("registering routes")
	// Endpoint to store a secret.
	app.Post("/secret", handler, func(c *fiber.Ctx) error {
		log.Infof("New secret request from %s", c.IP())
		// Limit the number of requests.
		if !limiter.Allow() {
			return HandleRateLimitError(c, fmt.Sprintf("Too many requests from %v, limit: %v", c.IP(), limiter.Limit()), nil)
		}

		type RequestBody struct {
			Secret string `json:"secret"`
		}
		var body RequestBody
		if err := c.BodyParser(&body); err != nil {
			return HandleValidationError(c, "Cannot parse JSON request body", err)
		}
		if body.Secret == "" {
			return HandleValidationError(c, "Secret is required", nil)
		}

		// Generate a unique key.
		uid := uuid.New().String()
		key, err := HideIdentifier(uid, []byte(encKey))
		if len(key) > keyLen {
			key = key[:keyLen]
		}
		if err != nil {
			return HandleServerError(c, "Failed to generate key", err)
		}

		// Insert the secret and key into the database.
		_, err = db.Exec("INSERT INTO secrets(key, secret) VALUES($1, $2)", key, body.Secret)
		if err != nil {
			return HandleDatabaseError(c, "Failed to store secret in database", err)
		}

		return c.JSON(fiber.Map{"key": key})
	})

	// Endpoint to retrieve a secret exactly once.
	app.Get("/secret/:key", handler, func(c *fiber.Ctx) error {
		// Limit the number of requests.
		if !limiter.Allow() {
			return HandleRateLimitError(c, "Too many requests", nil)
		}

		key := c.Params("key")

		// Start a transaction to ensure atomic read-update.
		tx, err := db.Begin()
		if err != nil {
			return HandleDatabaseError(c, "Failed to start database transaction", err)
		}
		defer func() {
			if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
				log.Error("Failed to rollback transaction", "error", err)
			}
		}()

		var secret string
		var retrievedAt *time.Time
		err = tx.QueryRow("SELECT secret, retrieved_at FROM secrets WHERE key = $1 FOR UPDATE", key).Scan(&secret, &retrievedAt)
		if err == sql.ErrNoRows {
			return HandleNotFoundError(c, fmt.Sprintf("Secret with key %s not found", key), nil)
		} else if err != nil {
			return HandleDatabaseError(c, "Failed to query secret from database", err)
		}

		// Check if the secret has already been retrieved.
		if retrievedAt != nil || secret == "" {
			return HandleNotFoundError(c, "Secret already retrieved", nil)
		}

		// Clear the secret and record the retrieval time.
		now := time.Now()
		_, err = tx.Exec("UPDATE secrets SET secret = '', retrieved_at = $1 WHERE key = $2", now, key)
		if err != nil {
			return HandleDatabaseError(c, "Failed to update secret in database", err)
		}

		// Commit the transaction.
		if err = tx.Commit(); err != nil {
			return HandleDatabaseError(c, "Failed to commit transaction", err)
		}

		// Return the original secret.
		return c.JSON(fiber.Map{"secret": secret})
	})
}
