.PHONY: dev test openapi drift-check sqlc tidy

dev:
	air

test:
	go test ./... -timeout=300s

openapi:
	swag init -g cmd/api/main.go -o docs --parseInternal --parseDependency

drift-check:
	$(MAKE) openapi
	@git diff --exit-code docs/ || (echo "ERROR: openapi docs are stale. Run 'make openapi' and commit." && exit 1)
	@git diff --exit-code internal/db/sqlc/ || (echo "ERROR: sqlc output is stale. Run 'make sqlc' and commit." && exit 1)

sqlc:
	sqlc generate

tidy:
	go mod tidy
