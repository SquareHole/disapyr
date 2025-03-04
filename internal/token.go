package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func GetAccessToken() (string, error) {

	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		return "", fmt.Errorf("error loading .env file: %w", err)
	}
	urlEnv := os.Getenv("URL")
	if urlEnv == "" {
		return "", fmt.Errorf("environment variable URL is not set")
	}
	url := fmt.Sprintf("https://%s/oauth/token", urlEnv)

	payload := strings.NewReader(fmt.Sprintf("{\"client_id\":\"%s\",\"client_secret\":\"%s\",\"audience\":\"%s\",\"grant_type\":\"%s\"}",
		os.Getenv("CLIENT_ID"),
		os.Getenv("CLIENT_SECRET"),
		os.Getenv("AUDIENCE"),
		os.Getenv("GRANT_TYPE"),
	))

	req, _ := http.NewRequest("POST", url, payload)

	req.Header.Add("content-type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("error unmarshalling response body: %w", err)
	}

	accessToken, ok := result["access_token"].(string)
	if !ok {
		return "", fmt.Errorf("access_token not found in response body")
	}

	return accessToken, nil
}
