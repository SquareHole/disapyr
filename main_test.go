package main

import (
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

var Limiter *rate.Limiter

func init() {
	// Load environment variables from .env file.
	if err := godotenv.Load(); err != nil {
		panic("Error loading .env file")
	}

	// Retrieve rate limit from environment variable.
	rateLimitStr := os.Getenv("RATE_LIMIT")
	rateLimit := 5 // Default to 5 requests per second.
	if rateLimitStr != "" {
		var err error
		rateLimit, err = strconv.Atoi(rateLimitStr)
		if err != nil {
			panic("Invalid RATE_LIMIT value:" + err.Error())
		}
	}

	// Initialize the Limiter variable.
	Limiter = createRateLimiter(rateLimit)
}

func TestRateLimiterWithDifferentRate(t *testing.T) {
	// Create a new rate limiter with a rate of 2 req/s and burst of 4.
	localLimiter := createRateLimiter(2)

	// Create a new Fiber app.
	app := fiber.New()

	// Define a test handler.
	app.Get("/test", func(c *fiber.Ctx) error {
		if !localLimiter.Allow() {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"error": "too many requests"})
		}
		return c.SendString("OK")
	})

	// Create a test request.
	req := httptest.NewRequest("GET", "/test", nil)

	// With a burst of 4, the first 4 requests should be allowed.
	for i := 0; i < 4; i++ {
		resp, _ := app.Test(req)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	}

	// The 5th immediate request should be rejected.
	resp, _ := app.Test(req)
	assert.Equal(t, fiber.StatusTooManyRequests, resp.StatusCode)

	// Optionally, wait for a second to let the limiter refill tokens.
	time.Sleep(time.Second)

	// After one second, you should have 2 tokens refilled (up to a max of 4).
	// So, the next 2 requests will be allowed.
	for i := 0; i < 2; i++ {
		resp, _ := app.Test(req)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	}

	// And a subsequent request without waiting might be rejected if it exceeds the available tokens.
	resp, _ = app.Test(req)
	// Depending on the current bucket state, assert the expected status.
	// For instance, if you expect the bucket to be empty, then:
	assert.Equal(t, fiber.StatusTooManyRequests, resp.StatusCode)
}

func TestRateLimiter(t *testing.T) {
	// Create a new Fiber app.
	app := fiber.New()

	// Define a test handler.
	app.Get("/test", func(c *fiber.Ctx) error {
		// Limit the number of requests.
		if !Limiter.Allow() {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"error": "too many requests"})
		}
		return c.SendString("OK")
	})

	// Create a test request.
	req := httptest.NewRequest("GET", "/test", nil)

	// Make requests within the rate limit.
	rateLimitStr := os.Getenv("RATE_LIMIT")
	rateLimit := 5 // Default to 5 requests per second.
	if rateLimitStr != "" {
		var err error
		rateLimit, err = strconv.Atoi(rateLimitStr)
		if err != nil {
			panic("Invalid RATE_LIMIT value:" + err.Error())
		}
	}
	burstSize := 2 * rateLimit
	for i := 0; i < burstSize; i++ {
		resp, _ := app.Test(req)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	}

	// Verify that the rate limit is exceeded.
	resp, _ := app.Test(req)
	assert.Equal(t, fiber.StatusTooManyRequests, resp.StatusCode)
}

func TestMain(m *testing.M) {
	// Run the tests.
	exitCode := m.Run()

	// Exit with the appropriate code.
	os.Exit(exitCode)
}
