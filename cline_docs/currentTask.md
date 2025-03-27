# Current Task: Implementing Service Layer

## Current Objectives
- Implement a service layer to separate business logic from HTTP handlers
- Create a repository layer to abstract database operations
- Improve code organization and maintainability
- Prepare for future architectural improvements

## Context
The Disapyr application currently has business logic mixed with HTTP handlers in the `internal/api.go` file. This makes the code harder to maintain, test, and extend. By implementing a service layer, we can separate the business logic from the HTTP handlers, making the code more modular and easier to test.

The service layer will be responsible for:
1. Implementing the business logic for storing and retrieving secrets
2. Coordinating between the HTTP handlers and the repository layer
3. Handling validation and error handling

The repository layer will be responsible for:
1. Abstracting database operations
2. Providing a clean interface for the service layer to interact with the database
3. Handling database-specific error handling and transactions

These improvements are documented in the projectRoadmap.md file under "Improve Architecture" and are the next priority items to address after completing the security improvements.

## Next Steps

### 1. Create Repository Layer
- Create a new file `internal/repository/secret_repository.go` to implement the repository layer
- Define interfaces for the repository layer
- Implement the repository layer for PostgreSQL
- Add unit tests for the repository layer

### 2. Implement Service Layer
- Create a new file `internal/service/secret_service.go` to implement the service layer
- Define interfaces for the service layer
- Implement the service layer using the repository layer
- Add unit tests for the service layer

### 3. Refactor API Handlers
- Update `internal/api.go` to use the service layer
- Remove business logic from HTTP handlers
- Ensure proper error handling and validation
- Add integration tests for the API handlers

### 4. Update Documentation
- Update the codebase documentation to reflect the new architecture
- Document the interfaces and their implementations
- Update the architecture diagram if necessary
