SHELL=/bin/bash

all: migrate api

db:
	docker run --name stock-db -e POSTGRES_PASSWORD=${POSTGRES_PASSWORD} -e POSTGRES_DB=${DATABASE_NAME} -p 5432 -d postgres:13-alpine

test_db:
	docker run --name stock-db-test -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=stock_db_test -p 5432:5432 -d postgres:13-alpine

migrate:
	@PORT="$(shell docker port stock-db 5432 | awk -F ':' '{print $$2}')" && sql-migrate up -limit 0

api:
	echo "Not implemented"

destroy:
	@docker stop stock-db
	@docker rm stock-db

clean:
	sql-migrate down -limit 0
