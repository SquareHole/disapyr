package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
)

func main() {
	// Define CLI flags.
	storeCmd := flag.Bool("store", false, "Store a secret")
	retrieveCmd := flag.Bool("retrieve", false, "Retrieve a secret")
	secretVal := flag.String("secret", "", "Secret to store (use with --store)")
	keyVal := flag.String("key", "", "Key to retrieve the secret (use with --retrieve)")
	server := flag.String("server", "https://localhost:3000", "API server URL (default: https://localhost:3000)")
	flag.Parse()

	// Ensure that either store or retrieve is selected, but not both.
	if *storeCmd && *retrieveCmd {
		fmt.Println("Error: Cannot use both --store and --retrieve flags together.")
		os.Exit(1)
	}
	if !*storeCmd && !*retrieveCmd {
		fmt.Println("Error: Please specify either --store or --retrieve flag.")
		flag.Usage()
		os.Exit(1)
	}

	if *storeCmd {
		// Validate input.
		if *secretVal == "" {
			fmt.Println("Error: Please provide a secret using the --secret flag.")
			os.Exit(1)
		}

		// Prepare and send a POST request to /secret.
		url := fmt.Sprintf("%s/secret", *server)
		payload := map[string]string{"secret": *secretVal}
		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			fmt.Println("Error marshalling JSON:", err)
			os.Exit(1)
		}

		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}
		resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
		if err != nil {
			fmt.Println("Error calling API:", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response:", err)
			os.Exit(1)
		}
		if resp.StatusCode != http.StatusOK {
			fmt.Printf("API error: %s\n", body)
			os.Exit(1)
		}

		// Display the returned key.
		fmt.Printf("Secret stored successfully. Key: %s\n", body)
	} else if *retrieveCmd {
		// Validate input.
		if *keyVal == "" {
			fmt.Println("Error: Please provide a key using the --key flag.")
			os.Exit(1)
		}

		// Prepare and send a GET request to /secret/:key.
		url := fmt.Sprintf("%s/secret/%s", *server, *keyVal)
		resp, err := http.Get(url)
		if err != nil {
			fmt.Println("Error calling API:", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response:", err)
			os.Exit(1)
		}
		if resp.StatusCode != http.StatusOK {
			fmt.Printf("API error: %s\n", body)
			os.Exit(1)
		}

		// Display the retrieved secret.
		fmt.Printf("Retrieved secret: %s\n", body)
	}
}
