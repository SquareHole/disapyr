# Disapyr

## Description
Disapyr is a web application that allows users to store a secret and generate a unique URL for retrieving that secret. The secret can be copied to the clipboard and reshared, generating a new unique URL. The application ensures that secrets are stored securely and are unrecoverable after a defined timeout or once retrieved.

## Installation
1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd disapyr
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

## Usage
1. Start the application:
   ```bash
   go run main.go
   ```

2. Open your browser and navigate to `http://localhost:3000`.

## Environment Variables
The following environment variables are required to run the application:

- `CLIENT_ID`: Client ID for external authentication
- `CLIENT_SECRET`: Client secret for external authentication
- `AUDIENCE`: Intended audience for tokens
- `GRANT_TYPE`: Grant type for authentication
- `DB_USER`: Database username
- `DB_PASSWORD`: Database password
- `DB_HOST`: Database host
- `DB_PORT`: Database port
- `DB_NAME`: Database name
- `ENC_KEY`: Encryption key
- `KEY_LEN`: Key length for encryption
- `RATE_LIMIT`: Maximum requests allowed
- `CERT_PATH`: Path to the certificate file
- `KEY_PATH`: Path to the key file
- `BASE_URL`: Base URL for the application
- `URL`: Domain URL for authentication
- `UI_HOST_PORT`: Port for the UI host
- `DB_USESSL`: Enable SSL for the database connection

## API Endpoints

### POST /secret
Stores a secret and returns a unique key.

**Request Body:**
```json
{
  "secret": "your_secret_here"
}
```

**Response:**
```json
{
  "key": "unique_key"
}
```

### GET /secret/:key
Retrieves the secret using the unique key.

**Response:**
```json
{
  "secret": "your_secret_here"
}
```

## CLI Usage
The CLI is the main application and can be built and run as follows:

1. Build the application:
   ```bash
   go build -o disapyr cmd/main.go
   ```

2. Run the application:
   ```bash
   ./disapyr
   ```

Note: You may need to set the environment variables before running the application.

### Usage Examples

1.  **Store a secret:**

    ```bash
    ./disapyr -store -secret "your_secret_here"
    ```

2.  **Retrieve a secret:**

    ```bash
    ./disapyr -retrieve -key "the_key_you_received"
    ```

    *   Replace `"the_key_you_received"` with the actual key provided when storing the secret.

## Certificate Generation
To generate a self-signed certificate for HTTPS, run the following command:

```bash
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes
```

To trust the self-signed certificate on your machine, run the following command:

```bash
sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain cert.pem
```

## License
This project is licensed under the MIT License.
