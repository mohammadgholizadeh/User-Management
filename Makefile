MIGRATIONS_PATH=./migrations/

POSTGRES_DB_NAME=user_management
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=1234
POSTGRES_DB_URL=postgresql://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB_NAME)?sslmode=disable
VERSION=1

POSTGRES_CONTAINER=postgres_user_management
REDIS_CONTAINER=redis_user_management

tidy:
	go mod tidy

fmt:
	go fmt ./...
	swag fmt --dir "./cmd/server"

lint:
	golangci-lint run --config "./config/.golangci.yaml"

swagger:
	swag init -g "./cmd/server/main.go" --parseDependency --output "./docs"

build-service:
	go build -o usermanagement ./cmd/server/

run-service:
	./usermanagement

postgres-container:
	docker run --name $(POSTGRES_CONTAINER) -p $(POSTGRES_PORT):5432 -e POSTGRES_USER=$(POSTGRES_USER) -e POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) -d postgres:alpine

redis-container:
	docker run --name $(REDIS_CONTAINER) -p 6379:6379 -e REDIS_PASSWORD=redis -d redis:alpine

start-containers:
	docker start $(POSTGRES_CONTAINER)
	docker start $(REDIS_CONTAINER)

stop-containers:
	docker stop $(POSTGRES_CONTAINER)
	docker stop $(REDIS_CONTAINER)

create-postgres-db:
	docker exec -it $(POSTGRES_CONTAINER) createdb --username=$(POSTGRES_USER) --owner=postgres $(POSTGRES_DB_NAME)

drop-postgres-db:
	docker exec -it $(POSTGRES_CONTAINER) dropdb --username=$(POSTGRES_USER) $(POSTGRES_DB_NAME)

migrateup:
	migrate -path "$(MIGRATIONS_PATH)" -database "$(POSTGRES_DB_URL)" -verbose up

migrateup-1:
	migrate -path "$(MIGRATIONS_PATH)" -database "$(POSTGRES_DB_URL)" -verbose up 1

migratedown:
	migrate -path "$(MIGRATIONS_PATH)" -database "$(POSTGRES_DB_URL)" -verbose down

migratedown-1:
	migrate -path "$(MIGRATIONS_PATH)" -database "$(POSTGRES_DB_URL)" -verbose down 1

create-migration:
	migrate create -ext sql -dir "$(MIGRATIONS_PATH)" -seq "$(MIGRATION_NAME)"

fix-migrate:
	migrate -database "$(POSTGRES_DB_URL)" -path "$(MIGRATIONS_PATH)" force $(VERSION)

pipeline: tidy swagger fmt build-service run-service

build-image:
	docker build --build-arg GITHUB_TOKEN=$(TOKEN) -t userservice-image -f Dockerfile .

run-app-container:
	docker run -e USER_POSTGRES_HOST=192.168.185.161 -e USER_REDIS_HOST=192.168.185.161 -e USER_HTTP_SERVER_IP=0.0.0.0 -e USER_GRPC_SERVER_IP=0.0.0.0 -e USER_BACKEND_URL=http://192.168.185.161/api/v1/user-mg --net=host --rm -d --name userservice userservice-image