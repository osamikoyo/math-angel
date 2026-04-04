# Math-Angel

A web application for managing and practicing math tasks with user ratings and caching.

## Prerequisites

- Go 1.21+
- Bun (for frontend)
- SQLite (database)
- Redis (cache)

## Installation

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd math-angel
   ```

2. Install Go dependencies:
   ```bash
   go mod tidy
   ```

3. Install frontend dependencies:
   ```bash
   bun install
   ```

4. Set up configuration in `config.yaml` (database path, Redis settings, etc.).

## Running

1. Build the project:
   ```bash
   task build
   ```

2. Run the application:
   ```bash
   task run
   ```

3. Run the application with watch functionality:
    ```bash
    task watch
    ```

The server will start on the address specified in `config.yaml` (default: localhost:8080).

## Development

- Watch mode: `task watch`
- Run tests: `go test ./...`

## Project Structure

- `cmd/`: Main application entry point
- `internal/`: Core business logic (app, service, repository, etc.)
- `static/`: Static assets (CSS, JS)
- `ui/`: Templ templates for HTML
