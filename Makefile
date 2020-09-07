SHELL := /bin/bash
VERSION := $(shell cat .version)

.PHONY: \
	help \
	run \
	help \
	clean \
	clean-artifacts \
	clean-releases \
	tools \
	test \
	vet \
	lint \
	fmt \
	env \
	build \
	doc \
	version

all: fmt lint vet build

help:
	@echo 'Usage: make <OPTIONS> ... <TARGETS>'
	@echo ''
	@echo 'Available targets are:'
	@echo ''
	@echo '    run                Run sic.'
	@echo '    help               Show this help screen.'
	@echo '    clean              Remove binaries, artifacts and releases.'
	@echo '    clean-artifacts    Remove build artifacts only.'
	@echo '    clean-releases     Remove releases only.'
	@echo '    tools              Install tools needed by the project.'
	@echo '    test               Run unit tests.'
	@echo '    vet                Run go vet.'
	@echo '    lint               Run golint.'
	@echo '    fmt                Run go fmt.'
	@echo '    env                Display Go environment.'
	@echo '    build              Build project for current platform.'
	@echo '    doc                Start Go documentation server on port 8080.'
	@echo '    version            Display Go version.'
	@echo ''
	@echo 'Targets run by default are: fmt, lint, vet, and build.'
	@echo ''

print-%:
	@echo $* = $($*)

clean: clean-artifacts clean-releases
	go clean -i ./...
	rm -vf \
	  $(CURDIR)/coverage.* \

clean-artifacts:
	rm -Rf artifacts/*

clean-releases:
	rm -Rf releases/*

clean-all: clean clean-artifacts

# Run tests
test: build
	go test ./cmd/... ./internal/... -coverprofile coverage.out

tools:
	go get -u golang.org/x/lint/golint
	go get github.com/axw/gocov/gocov
	go get github.com/AlekSi/gocov-xml

vet:
	go vet -v ./...

lint: tools
	golint -set_exit_status  ./...

fmt:
	go fmt ./...

env:
	@go env

# Run against the configured Kubernetes cluster in ~/.kube/configs
run: build
	go run ./cmd/main.go(TARGET)" .

build: fmt vet
	go build -o bin/sic ./cmd/root.go

doc:
	godoc -http=:8080 -index

version:
	@go version
