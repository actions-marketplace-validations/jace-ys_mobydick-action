include .env

.PHONY: run generate test

run:
	@go run main.go distribute \
		--organisation ${ORGANISATION} \
		--token ${TOKEN} \
		--private

generate:
	@go generate ./...

test:
	@go test -v ./...
