SHELL := /bin/bash
GOOS := $(shell go env GOOS)

# place binary in ./bin
build:
	go build -o bin/$(GOOS)/monictl cmd/monictl.go

# place binary in ~/go/bin
.PHONY: install
install:
	go install cmd/monictl.go

.PHONY: run
run:
	go run cmd/monictl.go

.PHONY: cross
cross:
	echo "Compiling for the following platforms: darwin, linux, windows"
	GOOS=darwin go build -o bin/darwin/monictl cmd/monictl.go
	GOOS=linux go build -o bin/linux/monictl cmd/monictl.go
	GOOS=windows go build -o bin/windows/monictl.exe cmd/monictl.go
.PHONY: fmt
fmt:
	go fmt ./...
.PHONY: tidy
tidy:
	go mod tidy
