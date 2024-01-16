install:
	go mod download

test:
	go clean --testcache
	go test ./...

run-service:
	go run cmd/service/main.go

run-client:
	go run cmd/client/main.go

run-prod:
	docker-compose up --abort-on-container-exit --force-recreate --build service --build client