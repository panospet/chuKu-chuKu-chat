SHELL := /bin/bash

CONFIG_FILE ?= .env
APP_DSN ?= $(shell cat .env | grep DSN | awk -FDSN= '{ print $$2}')
MIGRATE := docker run --rm -v $(shell pwd)/migrations:/migrations --network host --user $(id -u):$(id -g) migrate/migrate -path=/migrations/ -database "$(APP_DSN)"
MIGRATE_CREATE := docker run --rm -v $(shell pwd)/migrations:/migrations --network host --user $(shell id -u):$(shell id -g) migrate/migrate create --seq -ext sql -dir /migrations/
CWD := $(shell pwd)

.PHONY: default
default: help

# generate help info from comments: thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help
help: ## help information about make commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' 

.PHONY: run
run: ## run the app
	go generate cmd/app/main.go
	go run -tags=jsoniter cmd/app/main.go

.PHONY: db-start
db-start: ## start the database server
	@mkdir -p testdata/postgres
	docker run --rm --net host --name chukudb -d -v $(shell pwd)/testdata:/testdata \
		-v $(shell pwd)/testdata/postgres:/var/lib/postgresql/data \
		-e POSTGRES_PASSWORD=password -e POSTGRES_DB=chukudb -d postgres:12.2-alpine

.PHONY: db-stop
db-stop: ## stop the database server
	docker stop chukudb

.PHONY: db-login
db-login: ## login to the database
	docker exec -it chukudb psql -U postgres -d chukudb

.PHONY: migrate
migrate: ## run all new database migrations
	@echo "Running all new database migrations..."
	@$(MIGRATE) up

.PHONY: migrate-down
migrate-down: ## revert database to the last migration step
	@echo "Reverting database to the last migration step..."
	@$(MIGRATE) down 1

.PHONY: migrate-new
migrate-new: ## create a new database migration
	@read -p "Enter the name of the new migration: " name; \
	$(MIGRATE_CREATE) $${name}

.PHONY: redis-start
redis-start: ## start redis (no persistence)
	docker run --rm --name chuku-redis -d -p 6379:6379 redis

.PHONY: redis-stop
redis-stop: ## stop redis (no persistence)
	docker stop chuku-redis