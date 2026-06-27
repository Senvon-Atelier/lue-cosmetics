.PHONY: dev test openapi drift-check up down

up:
	docker compose up -d postgres mailpit

down:
	docker compose down

dev: up
	$(MAKE) -C backend dev

test:
	$(MAKE) -C backend test

openapi:
	$(MAKE) -C backend openapi

drift-check:
	$(MAKE) -C backend drift-check
