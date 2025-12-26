When you create a new file, run git add on it to stage it
Use Go best practices when creating new functions
Do not update any files in the HTTP_COLLECTION or _http_docs folders

## Project Structure

- `api/` - API endpoints and handlers
- `cmd/` - Application entry points
  - `service-atlas/` - Main application
- `internal/` - Private application code
  - `config/` - Configuration management
- `HTTP_COLLECTION/` - API request examples (Bruno format)
- `_http_docs/` - API documentation

## Tech Stack

- **Language**: Go 1.21+
- **Database**: Neo4j
- **HTTP Framework**: Standard Go HTTP package
- **Testing**: Go standard testing package
- **Containerization**: Docker and Docker Compose

## Testing
- whenever running tests, make sure to use the `--short` flag