package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	// GET handler to serve the main page for capturing the secret.
	app.Get("/", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(`
<html>
<head>
    <title>Secret Capture</title>
    <link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.5.2/css/bootstrap.min.css">
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <style>
        /* Reserve space so overlapping elements don’t shift layout */
        #contentContainer {
            position: relative;
            min-height: 250px;
        }
        /* Overlapping containers for input and result */
        #inputContainer, #resultContainer {
            position: absolute;
            width: 100%;
            top: 0;
            left: 0;
            transition: opacity 500ms;
        }
    </style>
    <script>
        // Helper to fade in an element.
        function fadeInElement(el, duration) {
            el.style.opacity = 0;
            el.style.display = 'block';
            el.style.transition = 'opacity ' + duration + 'ms';
            setTimeout(function() {
                el.style.opacity = 1;
            }, 10);
        }

        // Fade out the input container on submit.
        function handleSubmit(e) {
            var inputContainer = document.getElementById('inputContainer');
            inputContainer.style.transition = 'opacity 500ms';
            inputContainer.style.opacity = 0;
        }

        // After HTMX swaps in the response, fade in the result,
        // update the title, and add a "New secret" button.
        document.addEventListener("htmx:afterSwap", function(event){
            var resultContainer = document.getElementById('resultContainer');
            fadeInElement(resultContainer, 500);
            document.getElementById('pageTitle').textContent = 'One time link to share';
            if (!document.getElementById('newSecretBtn')) {
                var newBtn = document.createElement('button');
                newBtn.id = 'newSecretBtn';
                newBtn.textContent = 'New secret';
                newBtn.className = 'btn btn-secondary mt-3';
                newBtn.onclick = function() {
                    var resultContainer = document.getElementById('resultContainer');
                    resultContainer.style.transition = 'opacity 500ms';
                    resultContainer.style.opacity = 0;
                    setTimeout(function() {
                        resultContainer.innerHTML = '';
                        var inputContainer = document.getElementById('inputContainer');
                        inputContainer.style.opacity = 0;
                        inputContainer.style.display = 'block';
                        fadeInElement(inputContainer, 500);
                        document.getElementById('secret').value = '';
                        document.getElementById('pageTitle').textContent = 'Capture your secret';
                    }, 500);
                };
                resultContainer.appendChild(newBtn);
            }
        });
    </script>
</head>
<body>
    <div class="container d-flex justify-content-center align-items-center" style="height: 100vh;">
        <div class="text-center">
            <h1 id="pageTitle">Capture your secret</h1>
            <div id="contentContainer">
                <form id="secretForm" hx-post="/" hx-target="#resultContainer" hx-swap="innerHTML" onsubmit="handleSubmit(event)">
                    <div id="inputContainer">
                        <label for="secret">Secret:</label>
                        <textarea id="secret" name="secret" class="form-control mx-auto" rows="4" style="width: 300px;"></textarea><br>
                        <button type="submit" class="btn btn-primary">Capture</button>
                    </div>
                </form>
                <div id="resultContainer" style="opacity:0;"></div>
            </div>
        </div>
    </div>
</body>
</html>
		`)
	})

	// POST handler to capture the secret and generate the one-time link.
	app.Post("/", func(c *fiber.Ctx) error {
		secret := c.FormValue("secret")
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
				return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
			}
			defer resp.Body.Close()

			var apiResponse struct {
				Key string `json:"key"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
				return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
			}
			// The external API returns a key which is used to build the one-time link.
			url := fmt.Sprintf("https://localhost:3000/secret/%s", apiResponse.Key)
			c.Set("Content-Type", "text/html; charset=utf-8")
			return c.SendString(fmt.Sprintf("<div>%s</div>", url))
		}
		return c.SendString("No secret provided")
	})

	// GET handler for displaying a secret retrieved from the external API.
	app.Get("/secret/:key", func(c *fiber.Ctx) error {
		key := c.Params("key")
		apiURL := fmt.Sprintf("https://localhost:3000/secret/%s", key)

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}
		resp, err := client.Get(apiURL)
		if err != nil {
			log.Printf("Error during API GET call: %v", err)
			return displaySecretPage(c, "Error retrieving secret")
		}
		defer resp.Body.Close()

		// If the API returns a non-200 status, display a message in the readonly textbox.
		if resp.StatusCode != http.StatusOK {
			message := "Secret not found. It may have already been retrieved."
			return displaySecretPage(c, message)
		}

		var apiResponse struct {
			Secret string `json:"secret"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}

		return displaySecretPage(c, apiResponse.Secret)
	})

	fmt.Println("Starting server on :8080")
	log.Fatal(app.Listen(":8080"))
}

// displaySecretPage renders an HTML page with a read-only textarea containing the provided content.
func displaySecretPage(c *fiber.Ctx, content string) error {
	html := fmt.Sprintf(`
<html>
<head>
    <title>Your Secret</title>
    <link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.5.2/css/bootstrap.min.css">
    <script>
        function copySecret() {
            var copyText = document.getElementById("secret");
            copyText.select();
            copyText.setSelectionRange(0, 99999); // For mobile devices
            document.execCommand("copy");
            alert("Copied the secret");
        }
    </script>
</head>
<body>
    <div class="container d-flex justify-content-center align-items-center" style="height: 100vh;">
        <div class="text-center">
            <h1>Your Secret</h1>
            <textarea id="secret" class="form-control" rows="4" style="width: 300px;" readonly>%s</textarea>
            <br>
            <button class="btn btn-primary" onclick="copySecret()">Copy Secret</button>
        </div>
    </div>
</body>
</html>
`, content)
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.SendString(html)
}
