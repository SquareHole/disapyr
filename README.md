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

- `DB_USER`: Database username
- `DB_PASSWORD`: Database password
- `DB_HOST`: Database host
- `DB_PORT`: Database port
- `DB_NAME`: Database name
- `ENC_KEY`: Encryption key
- `KEY_LEN`: Key length

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

## License
This project is licensed under the MIT License.