.PHONY: gen run test lint docker-build

gen:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/proto/v1/limiter.proto

run:
	go run cmd/server/main.go

test:
	go test -v ./...

lint:
	golangci-lint run

docker-build:
	docker build -t sanguis-server -f deployments/Dockerfile .
