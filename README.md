# User Management Service

Simple user management service built with Go, Echo, PostgreSQL, Redis, and Helm manifests for Kubernetes deployment.

## Prerequisites
- Go 1.24+
- PostgreSQL and Redis instances (local or remote)
- Docker (for container builds)
- Helm 3 (for Kubernetes deployment)

## Local Development
```bash
go run ./cmd/cli serve
```

## Database Seed
```bash
go run ./cmd/cli seed
```

## Docker Build
```bash
docker build -t user-management:local .
```

### Database Management
```bash
make postgres-container  # Start PostgreSQL container
make redis-container     # Start Redis container
make start-containers    # Start both containers
make stop-containers     # Stop both containers
make create-postgres-db  # Create database
make drop-postgres-db    # Drop database
make migrateup          # Run all migrations
make migratedown        # Rollback all migrations
make create-migration MIGRATION_NAME=my_migration  # Create new migration
```

### Deployment
```bash
make pipeline      # Complete build pipeline (tidy, swagger, fmt, build, run)
make build-image   # Build Docker image
make run-app-container  # Run app in container
```
