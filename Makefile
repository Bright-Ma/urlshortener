# migrate
install_migrate:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# sqlc
install_sqlc:
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# postgres
lanch_postgres:
	docker run --name postgres_urls \
	-e POSTGRES_USER=lang \
	-e POSTGRES_PASSWORD=password \
	-e POSTGRES_DB=urldb \
	-p 5432:5432 \
	-d postgres

# redis
lanch_redis:
	docker run --name=reids_urls \
	-p 6379:6379 \
	-d redis

databaseURL="postgres://lang:password@localhost:5432/urldb?sslmode=disable"

migrate_up:
	migrate -path="./database/migrate" -database=${databaseURL} up

migrate_drop:
	migrate -path="./database/migrate" -database=${databaseURL} drop -f