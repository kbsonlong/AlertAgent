# Technology Stack

## Backend Stack

- **Language**: Go 1.21+
- **Web Framework**: Gin (HTTP router and middleware)
- **ORM**: GORM (database abstraction layer)
- **Database**: MySQL 8.0+ (primary storage)
- **Cache**: Redis 6.0+ (caching and task queue)
- **Logging**: Zap (structured logging)
- **Configuration**: YAML-based config with hot reload via fsnotify
- **AI Integration**: Ollama (local AI model inference)

## Frontend Stack

- **Framework**: React 19.0+ with TypeScript
- **Build Tool**: Vite 6.2+
- **UI Library**: Ant Design 5.24+
- **HTTP Client**: Axios 1.8+
- **Routing**: React Router DOM 7.4+
- **Icons**: Ant Design Icons 6.0+

## Development Tools

- **Package Management**: Go modules, npm
- **Code Quality**: ESLint, golangci-lint
- **Hot Reload**: Air (Go), Vite (React)
- **Container**: Docker & Docker Compose for development environment

## Common Commands

### Development Environment

```bash
# Start local development environment
make dev

# Start Docker development environment  
make docker-dev

# Stop development environment
make dev-stop
make docker-dev-stop

# Check environment setup
make check
```

### Project Management

```bash
# Install dependencies
make deps

# Build project
make build

# Run tests
make test

# Code linting
make lint

# Clean build artifacts
make clean
```

### Backend Development

```bash
# Run Go application
go run cmd/main.go

# Install Go dependencies
go mod download && go mod tidy

# Run Go tests
go test -v ./...
```

### Frontend Development

```bash
# Start React dev server
cd web && npm run dev

# Install frontend dependencies
cd web && npm install

# Build frontend
cd web && npm run build

# Lint frontend code
cd web && npm run lint
```

## Configuration

- **Backend Config**: `config/config.yaml` (hot-reloadable)
- **Frontend Config**: Environment variables in `web/.env`
- **Docker Config**: `docker-compose.dev.yml` for development services