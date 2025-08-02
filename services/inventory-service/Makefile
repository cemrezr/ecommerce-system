DB_URL=postgres://user:pass@localhost:5433/inventory?sslmode=disable
MIGRATIONS_DIR=internal/migrations

run:
	go run cmd/main.go

migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up

migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down 1

migrate-reset:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up

migrate-force:
ifndef V
	$(error You must provide a version with V=... (e.g., make migrate-force V=0))
endif
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" force $(V)


replay:
	go run cmd/replay.go
