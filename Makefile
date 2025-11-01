createdb:
	docker exec -it simple-bank-db createdb --username=root --owner=root simple-bank

dropdb:
	docker exec -it simple-bank-db dropdb simple-bank

migrateup:
	migrate -path db/migration -database "postgresql://root:postgres@localhost:5432/simple-bank?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://root:postgres@localhost:5432/simple-bank?sslmode=disable" -verbose down

sqlc:
	sqlc generate

test: 
	go test -v -cover ./...

server:
	go run main.go

.PHONY: createdb dropdb migrateup migratedown sqlc server test