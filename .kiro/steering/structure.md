# Project Structure

## Root Directory Layout

```
alert_agent/
├── cmd/                    # Application entry points
├── config/                 # Configuration files
├── internal/              # Private application code
├── web/                   # Frontend React application
├── docs/                  # Documentation
├── scripts/               # Development and deployment scripts
├── templates/             # Template files
└── bin/                   # Compiled binaries
```

## Backend Structure (`internal/`)

### Core Directories

- **`api/v1/`** - HTTP handlers and API endpoints
- **`service/`** - Business logic layer
- **`model/`** - Data models and database schemas
- **`config/`** - Configuration management
- **`router/`** - Route definitions and middleware setup

### Infrastructure (`pkg/`)

- **`database/`** - Database connection and initialization
- **`logger/`** - Logging configuration and utilities
- **`queue/`** - Redis-based task queue implementation
- **`redis/`** - Redis client configuration
- **`types/`** - Shared type definitions

### Architecture Patterns

- **Layered Architecture**: API → Service → Model → Database
- **Dependency Injection**: Services receive dependencies via constructors
- **Error Handling**: Structured error types with context
- **Configuration**: Hot-reloadable YAML configuration

## Frontend Structure (`web/src/`)

### Core Directories

- **`pages/`** - Page components organized by feature
  - `alerts/` - Alert management pages
  - `knowledge/` - Knowledge base pages
  - `notifications/` - Notification management
  - `providers/` - Provider configuration
  - `settings/` - System settings
- **`components/`** - Reusable UI components
- **`services/`** - API client functions
- **`utils/`** - Utility functions and helpers
- **`layouts/`** - Layout components

### Frontend Patterns

- **Feature-based Organization**: Pages grouped by business domain
- **Service Layer**: Centralized API communication
- **TypeScript**: Strong typing for better development experience
- **Component Composition**: Reusable components with clear interfaces

## Configuration Files

- **`config/config.yaml`** - Main application configuration
- **`docker-compose.dev.yml`** - Development environment services
- **`Makefile`** - Development workflow automation
- **`go.mod`** - Go module dependencies
- **`web/package.json`** - Frontend dependencies

## Key Conventions

### Go Code Style

- Package names: lowercase, single word
- Interface names: end with 'er' (e.g., `Handler`, `Service`)
- Error variables: start with 'Err' (e.g., `ErrNotFound`)
- Constants: CamelCase for exported, camelCase for unexported

### API Conventions

- RESTful endpoints: `/api/v1/resource`
- HTTP methods: GET (read), POST (create), PUT (update), DELETE (remove)
- Response format: `{"code": 200, "msg": "success", "data": {...}}`
- Error responses: Include error code, message, and optional details

### Database Conventions

- Table names: snake_case, plural (e.g., `alert_rules`)
- Column names: snake_case (e.g., `created_at`)
- Primary keys: `id` (auto-increment)
- Timestamps: `created_at`, `updated_at`

### Frontend Conventions

- Component names: PascalCase (e.g., `AlertList.tsx`)
- File organization: Feature-based grouping
- API services: One file per domain (e.g., `alert.ts`)
- Type definitions: Interfaces for data structures