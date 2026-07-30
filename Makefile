.PHONY: help \
	backend-run backend-build backend-test backend-lint backend-tidy backend-seed \
	backend-migrate-up backend-migrate-down backend-migrate-create \
	frontend-install frontend-start frontend-ios frontend-android frontend-test frontend-lint \
	dev install

BACKEND_DIR := backend
FRONTEND_DIR := mobile

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-24s\033[0m %s\n", $$1, $$2}'

## --- Backend (Go, in ./$(BACKEND_DIR)) ---

backend-run: ## Run the Go API server
	cd $(BACKEND_DIR) && go run ./cmd/api

backend-build: ## Build the Go API binary
	cd $(BACKEND_DIR) && go build -o bin/api ./cmd/api

backend-test: ## Run backend tests
	cd $(BACKEND_DIR) && go test ./...

backend-lint: ## Lint backend code
	cd $(BACKEND_DIR) && go vet ./...

backend-tidy: ## Tidy go.mod/go.sum
	cd $(BACKEND_DIR) && go mod tidy

backend-migrate-up: ## Apply DB migrations (reads DATABASE_URL from backend/.env)
	cd $(BACKEND_DIR) && set -a && . ./.env && set +a && migrate -path internal/db/migrations -database "$$DATABASE_URL" up

backend-migrate-down: ## Roll back the last DB migration (reads DATABASE_URL from backend/.env)
	cd $(BACKEND_DIR) && set -a && . ./.env && set +a && migrate -path internal/db/migrations -database "$$DATABASE_URL" down 1

backend-migrate-create: ## Create a new migration: make backend-migrate-create name=add_users_table
	cd $(BACKEND_DIR) && migrate create -ext sql -dir internal/db/migrations -seq $(name)

backend-seed: ## Populate the database with the Arabic Basics starter course
	cd $(BACKEND_DIR) && go run ./cmd/seed

## --- Frontend (React Native/Expo, in ./$(FRONTEND_DIR)) ---

frontend-install: ## Install frontend dependencies
	cd $(FRONTEND_DIR) && npm install

frontend-start: ## Start the Expo dev server
	cd $(FRONTEND_DIR) && npx expo start

frontend-ios: ## Start the Expo dev server and open iOS simulator
	cd $(FRONTEND_DIR) && npx expo start --ios

frontend-android: ## Start the Expo dev server and open Android emulator
	cd $(FRONTEND_DIR) && npx expo start --android

frontend-test: ## Run frontend tests
	cd $(FRONTEND_DIR) && npm test

frontend-lint: ## Lint frontend code
	cd $(FRONTEND_DIR) && npm run lint

## --- Combined ---

install: backend-tidy frontend-install ## Install/tidy dependencies for both apps

dev: ## Run backend in the background and start the Expo dev server in the foreground
	@echo "Starting backend in background (logs: /tmp/quranlingo-backend.log)..."
	@cd $(BACKEND_DIR) && (go run ./cmd/api > /tmp/quranlingo-backend.log 2>&1 &)
	$(MAKE) frontend-start
