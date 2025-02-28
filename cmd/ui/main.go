package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			secret := r.FormValue("secret")
			if secret != "" {
				apiURL := "https://localhost:3000/secret"
				jsonData := []byte(fmt.Sprintf(`{"secret":"%s"}`, secret))

				tr := &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				}
				client := &http.Client{Transport: tr}
				resp, err := client.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
				if err != nil {
					log.Printf("Error during API call: %v", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}
				defer resp.Body.Close()

				var apiResponse struct {
					Key string `json:"key"`
				}
				if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				url := fmt.Sprintf("https://localhost:3000/secret/%s", apiResponse.Key)
				fmt.Fprintf(w, "<div id=\"result\">%s</div>", url)
				return
			}
		}

		fmt.Fprint(w, `
<html>
<head>
    <title>Secret Capture</title>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
</head>
<body>
    <div class="container">
        <h1>Capture your secret</h1>
        <form hx-post="/" hx-target="#result" hx-swap="innerHTML">
            <label for="secret">Secret:</label>
            <input type="text" id="secret" name="secret"><br><br>
            <button type="submit">Capture</button>
        </form>
        <div id="result"></div>
    </div>
</body>
</html>
`)
	})

	fmt.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
