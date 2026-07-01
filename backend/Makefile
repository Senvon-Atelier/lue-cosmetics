.PHONY: dev test openapi drift-check sqlc tidy seed-run seed-clean

dev:
	air

test:
	go test ./... -timeout=600s

openapi:
	$(shell go env GOPATH)/bin/swag init -g cmd/api/main.go -o docs --parseInternal --parseDependency

drift-check:
	$(MAKE) openapi
	$(MAKE) sqlc
	@git diff --exit-code docs/ || (echo "ERROR: openapi docs are stale. Run 'make openapi' and commit." && exit 1)
	@git diff --exit-code internal/db/sqlc/ || (echo "ERROR: sqlc output is stale. Run 'make sqlc' and commit." && exit 1)

sqlc:
	sqlc generate

tidy:
	go mod tidy

seed-run:
	@echo "Building seed binary..."
	@go build -o ./seed-bin ./cmd/seed
	@echo "Running seed..."
	@./seed-bin --data-dir="./cmd/seed/data" --db-url="${DATABASE_URL}"
	@rm -f ./seed-bin
	@echo "✅ Seed complete!"

seed-clean:
	@rm -f ./seed-bin
	@echo "🧹 Seed binary cleaned"
