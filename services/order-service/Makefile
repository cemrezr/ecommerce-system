include .env
export

MIGRATE_CMD = migrate -path migrations -database $(DB_DSN)

.PHONY: up down version create force drop run build

up:
	$(MIGRATE_CMD) up

down:
	$(MIGRATE_CMD) down

drop:
	$(MIGRATE_CMD) drop -f

version:
	$(MIGRATE_CMD) version

force:
ifndef version
	$(error "You must provide a version using: make force version=XXXXXXX")
endif
	$(MIGRATE_CMD) force $(version)

create:
ifndef name
	$(error "You must provide a name using: make create name=create_orders_table")
endif
	migrate create -ext sql -dir migrations -seq $(name)

run:
	go run cmd/main.go

build:
	go build -o bin/order-service cmd/main.go

replay:
	go run cmd/replay.go
