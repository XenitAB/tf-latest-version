lint:
	golangci-lint run ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

test: fmt vet
	go test --cover ./...

build:
	go build -o bin/tf-latest-version
