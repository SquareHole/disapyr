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

## License

## API Endpoints

### POST /store
Stores a secret and returns a unique URL.

**Request Body:**
```json
{
  "secret": "your_secret_here"
}
```

**Response:**
```json
{
  "url": "http://localhost:3000/retrieve/unique_id"
}
```

### GET /retrieve/:id
Retrieves the secret using the unique URL.

**Response:**
```json
{
  "secret": "your_secret_here"
}
```
This project is licensed under the MIT License.
