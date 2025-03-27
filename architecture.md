# Project Architecture Document

## Overview
The project is a secure secret management system that allows users to store and retrieve secrets via a web interface and a command-line interface (CLI). It uses a combination of the Fiber web framework, PostgreSQL for data storage, and JWT for authentication.

## Key Components

1. **Main Application (`main.go`)**
   - **Purpose**: Serves as the entry point for the API server.
   - **Functionality**:
     - Loads environment variables and configures logging.
     - Initializes the Fiber app, database connection, and rate limiter.
     - Registers routes for API endpoints.
     - Supports both HTTP and HTTPS.

2. **Command-Line Interface (`cmd/cli/main.go`)**
   - **Purpose**: Provides a CLI tool for storing and retrieving secrets.
   - **Functionality**:
     - Supports `--store` and `--retrieve` operations.
     - Interacts with the API server to store and retrieve secrets.
     - Handles input validation and error reporting.

3. **User Interface (`cmd/ui/main.go`)**
   - **Purpose**: Provides a web interface for capturing and displaying secrets.
   - **Functionality**:
     - Serves static files and handles GET and POST requests.
     - Interacts with an external API to store and retrieve secrets.
     - Displays one-time links for secret retrieval.

4. **Core Logic (`internal/api.go`)**
   - **Purpose**: Implements core functionality for API interactions.
   - **Functionality**:
     - Validates JWT tokens using Auth0.
     - Encrypts identifiers using AES-GCM.
     - Manages database connections and ensures the `secrets` table exists.
     - Implements rate limiting for API requests.
     - Registers routes for storing and retrieving secrets.

## Data Flow
- **Secret Storage**: Secrets are stored via the CLI or UI, which send requests to the API server. The server generates a unique key, encrypts it, and stores the secret in the database.
- **Secret Retrieval**: Secrets are retrieved using the unique key. The server ensures the secret is only retrieved once and clears it from the database after retrieval.

## External Dependencies
- **Fiber**: Used for creating the web server.
- **PostgreSQL**: Used for data storage.
- **JWT**: Used for authentication.
- **Auth0**: Used for token validation.
- **godotenv**: Used for loading environment variables.

## Security Considerations
- **HTTPS Support**: The application supports HTTPS to ensure secure communication.
- **Token Validation**: JWT tokens are validated using Auth0 to ensure secure access.
- **Encryption**: Identifiers are encrypted using AES-GCM to protect sensitive data.
