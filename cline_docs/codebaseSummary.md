# Codebase Summary

## Project Structure

The Disapyr application is organized into the following structure:

```
disapyr/
├── bin/                  # Compiled binaries
├── cmd/                  # Command-line applications
│   ├── cli/              # CLI for storing and retrieving secrets
│   │   └── main.go       # CLI entry point
│   └── ui/               # Web UI for the application
│       ├── main.go       # UI server entry point
│       ├── capture_secret.html  # HTML template for capturing secrets
│       ├── display_secret.html  # HTML template for displaying secrets
│       └── styles.css    # CSS styles for the UI
├── internal/             # Internal packages not meant for external use
│   ├── api.go            # Core API functionality
│   ├── api_test.go       # Tests for API functionality
│   └── token.go          # Token handling utilities
├── main.go               # Main API server entry point
├── main_test.go          # Tests for the main application
├── go.mod                # Go module definition
├── go.sum                # Go module checksums
├── Makefile              # Build automation
├── .gitignore            # Git ignore file
├── LICENSE               # License file
└── README.md             # Project documentation
```

## Key Components and Their Interactions

### 1. API Server (main.go)
- Entry point for the API server
- Initializes the Fiber app, database connection, and rate limiter
- Registers routes for API endpoints
- Configures HTTPS if enabled

### 2. Core API Logic (internal/api.go)
- Implements JWT token validation
- Provides encryption utilities for identifiers
- Manages database connections
- Implements rate limiting
- Defines API routes for storing and retrieving secrets

### 3. Token Handling (internal/token.go)
- Retrieves access tokens from Auth0
- Used by the UI server to authenticate API requests

### 4. CLI Client (cmd/cli/main.go)
- Provides a command-line interface for storing and retrieving secrets
- Communicates with the API server
- Handles command-line arguments and flags

### 5. Web UI Server (cmd/ui/main.go)
- Serves the web interface for the application
- Communicates with the API server
- Handles form submissions and displays results

### 6. Web UI Templates
- capture_secret.html: Form for entering secrets
- display_secret.html: Page for displaying retrieved secrets
- styles.css: Styling for the UI

## Data Flow

1. **Secret Storage**:
   - User enters a secret via CLI or web UI
   - Client sends a POST request to the API server
   - API server generates a unique key, encrypts it, and stores the secret in the database
   - API server returns the key to the client
   - Client displays the key to the user

2. **Secret Retrieval**:
   - User provides a key via CLI or web UI
   - Client sends a GET request to the API server
   - API server looks up the secret in the database
   - If found and not yet retrieved, API server returns the secret and marks it as retrieved
   - Client displays the secret to the user

## External Dependencies

### Core Dependencies
- **github.com/gofiber/fiber/v2**: Web framework for the API and UI servers
- **github.com/lib/pq**: PostgreSQL driver for database connectivity
- **github.com/golang-jwt/jwt/v5**: JWT parsing and validation
- **github.com/joho/godotenv**: Environment variable loading from .env files
- **github.com/charmbracelet/log**: Structured logging
- **golang.org/x/time/rate**: Rate limiting implementation

### Security Dependencies
- **crypto/aes**: AES encryption for identifiers
- **crypto/cipher**: GCM mode for authenticated encryption
- **github.com/mr-tron/base58**: Base58 encoding for URL-friendly strings

### Testing Dependencies
- **github.com/stretchr/testify/assert**: Assertion utilities for testing

## Recent Significant Changes

The codebase has recently undergone a comprehensive code review that identified several areas for improvement:

1. **Security Concerns**:
   - JWT token validation needs enhancement
   - TLS verification bypass in development mode
   - Potential information leakage in error messages
   - Lack of database connection pooling and timeouts

2. **Architectural Improvements**:
   - Need for a service layer to separate business logic from HTTP handlers
   - Need for a repository layer to abstract database operations
   - Need for a more robust configuration system
   - Need for request logging middleware

3. **Code Quality Enhancements**:
   - Inconsistent error handling patterns
   - Hardcoded values that should be configurable
   - Limited test coverage
   - Need for better code documentation

## User Feedback Integration

No user feedback has been integrated yet, as the application is still in development. Future iterations will incorporate user feedback on:

1. User interface usability
2. Error message clarity
3. Feature requests (e.g., secret expiration options)
4. Performance and reliability
