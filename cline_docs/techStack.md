# Technology Stack

## Backend

### Programming Language
- **Go (Golang)**: The entire application is written in Go, leveraging its strong typing, concurrency features, and performance characteristics.

### Web Framework
- **Fiber**: A fast, Express-inspired web framework for Go that provides a robust set of features for building web applications.
  - Used for both the API server and the UI server
  - Provides routing, middleware support, and request/response handling

### Database
- **PostgreSQL**: A powerful, open-source relational database system.
  - Stores secrets with their unique keys
  - Tracks retrieval status with timestamps
  - Uses transactions for atomic operations

### Authentication
- **JWT (JSON Web Tokens)**: Used for API authentication.
  - Tokens are validated using Auth0's JWKS endpoint
  - RSA signature verification is implemented
  - Current implementation needs enhancement for audience and issuer validation

### Encryption
- **AES-GCM**: Used for encrypting identifiers.
  - Provides authenticated encryption
  - Uses a random nonce for each encryption operation
  - Encrypted data is Base58-encoded for URL-friendly representation

## Frontend

### UI Framework
- **Bootstrap**: Used for responsive design and UI components.
  - Provides consistent styling across the application
  - Ensures mobile responsiveness

### JavaScript Libraries
- **htmx**: Used for AJAX requests and DOM updates without writing custom JavaScript.
  - Enables dynamic content updates
  - Simplifies form submissions and response handling

### Fonts
- **Google Fonts**: Bitter and Roboto fonts for typography.

## DevOps & Infrastructure

### Environment Configuration
- **godotenv**: Used for loading environment variables from .env files.
  - Simplifies configuration management
  - Allows different configurations for development and production

### Logging
- **charmbracelet/log**: Used for structured logging.
  - Supports JSON formatting
  - Includes timestamps and caller information
  - Configurable log levels

### Rate Limiting
- **golang.org/x/time/rate**: Implements token bucket algorithm for rate limiting.
  - Prevents abuse of the API
  - Configurable rate and burst limits

### TLS/HTTPS
- **Native Go TLS**: Support for HTTPS connections.
  - Configurable through environment variables
  - Currently bypasses verification in development mode (needs improvement)

## Architecture Decisions

### 1. Separation of CLI and UI
- The application provides both a command-line interface and a web interface
- Both interfaces interact with the same API endpoints
- This separation allows for flexibility in how users interact with the application

### 2. One-Time Secret Access
- Secrets are designed to be retrieved exactly once
- After retrieval, the secret is cleared from the database
- This ensures that sensitive information is not persistently stored

### 3. Environment-Based Configuration
- All configuration is managed through environment variables
- This allows for easy deployment across different environments
- A future improvement would be a more structured configuration system

### 4. Stateless API Design
- The API is designed to be stateless
- Each request contains all the information needed to process it
- This allows for horizontal scaling of the API server

### 5. Transaction-Based Database Operations
- Critical database operations use transactions
- This ensures data consistency, especially for the one-time secret retrieval
