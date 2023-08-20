build:
	@go build -o bin/bankAPI

run: build
	@./bin/bankAPI

test:
	@go test -v ./...