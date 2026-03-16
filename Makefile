build:
	@go build -o go-http-lib 

run: build
	@./bin/fs

test:
	@go test ./... -v