#!make
.SILENT:

run: server client

server:
	docker compose build server
	docker compose up -d --force-recreate server

client:
	docker compose build client
	docker compose up -d --force-recreate client

test:
	go clean --testcache
	go test ./...

deps:
	go mod download && go mod tidy
