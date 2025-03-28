package internal

import (
	"github.com/charmbracelet/log"
	"github.com/gofiber/fiber/v2"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// ErrorCategory defines the general category of an error
type ErrorCategory string

const (
	// AuthError represents authentication and authorization errors
	AuthError ErrorCategory = "auth"
	// ValidationError represents input validation errors
	ValidationError ErrorCategory = "validation"
	// DatabaseError represents database-related errors
	DatabaseError ErrorCategory = "database"
	// ServerError represents internal server errors
	ServerError ErrorCategory = "server"
	// RateLimitError represents rate limiting errors
	RateLimitError ErrorCategory = "rate_limit"
	// NotFoundError represents resource not found errors
	NotFoundError ErrorCategory = "not_found"
)

// ErrorStatusMap maps error categories to HTTP status codes
var ErrorStatusMap = map[ErrorCategory]int{
	AuthError:       fiber.StatusUnauthorized,
	ValidationError: fiber.StatusBadRequest,
	DatabaseError:   fiber.StatusInternalServerError,
	ServerError:     fiber.StatusInternalServerError,
	RateLimitError:  fiber.StatusTooManyRequests,
	NotFoundError:   fiber.StatusNotFound,
}

// ErrorMessageMap maps error categories to user-friendly error messages
var ErrorMessageMap = map[ErrorCategory]string{
	AuthError:       "Authentication failed",
	ValidationError: "Invalid request data",
	DatabaseError:   "Database operation failed",
	ServerError:     "Internal server error",
	RateLimitError:  "Too many requests",
	NotFoundError:   "Resource not found",
}

// HandleError logs an error with detailed information and returns a standardized error response
// This function should be used for all error handling in API handlers
func HandleError(c *fiber.Ctx, category ErrorCategory, logMessage string, err error) error {
	// Log the detailed error for debugging
	if err != nil {
		log.Error(logMessage, "error", err, "category", category)
	} else {
		log.Error(logMessage, "category", category)
	}

	// Get the appropriate status code and user-friendly message
	statusCode := ErrorStatusMap[category]
	message := ErrorMessageMap[category]

	// Return a standardized error response
	return c.Status(statusCode).JSON(ErrorResponse{Error: message})
}

// HandleAuthError is a convenience function for handling authentication errors
func HandleAuthError(c *fiber.Ctx, logMessage string, err error) error {
	return HandleError(c, AuthError, logMessage, err)
}

// HandleValidationError is a convenience function for handling validation errors
func HandleValidationError(c *fiber.Ctx, logMessage string, err error) error {
	return HandleError(c, ValidationError, logMessage, err)
}

// HandleDatabaseError is a convenience function for handling database errors
func HandleDatabaseError(c *fiber.Ctx, logMessage string, err error) error {
	return HandleError(c, DatabaseError, logMessage, err)
}

// HandleServerError is a convenience function for handling server errors
func HandleServerError(c *fiber.Ctx, logMessage string, err error) error {
	return HandleError(c, ServerError, logMessage, err)
}

// HandleRateLimitError is a convenience function for handling rate limit errors
func HandleRateLimitError(c *fiber.Ctx, logMessage string, err error) error {
	return HandleError(c, RateLimitError, logMessage, err)
}

// HandleNotFoundError is a convenience function for handling not found errors
func HandleNotFoundError(c *fiber.Ctx, logMessage string, err error) error {
	return HandleError(c, NotFoundError, logMessage, err)
}
