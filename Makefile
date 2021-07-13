SHELL := /bin/bash
VERSION := $(shell cat .version)
SPEC_VERSION ?= $(shell cat .version)

GO_FILES?=$$(find . -name '*.go' |grep -v vendor)
TAG?=latest

.GIT_COMMIT=$(shell git rev-parse HEAD)
.GIT_BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
.GIT_COMMIT_AND_BRANCH=$(.GIT_COMMIT)-$(subst /,-,$(.GIT_BRANCH))
.GIT_VERSION=$(shell git describe --tags 2>/dev/null || echo "$(.GIT_COMMIT)")
.FS_VERSION=$(shell cat .version)
.BUILDTIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
CHANGELOG_VERSION_ORIG=$(grep -m1 \## CHANGELOG.MD | sed -e "s/\].*\$//" |sed -e "s/^.*\[//")
CHANGELOG_VERSION=$(shell grep -m1 \ \[[0-9]*.[0-9]*.[0-9]*\] CHANGELOG.MD | sed -e "s/\].*$$//" |sed -e "s/^.*\[//")
BUILD_DIR ?= $(PWD)/dist/rpmbuild
SPEC_NAME ?= cray-site-init
SPEC_FILE ?= ${SPEC_NAME}.spec
SOURCE_NAME ?= ${SPEC_NAME}-${SPEC_VERSION}
SOURCE_PATH := ${BUILD_DIR}/SOURCES/${SOURCE_NAME}.tar.bz2
BUILD_METADATA ?= "$(shell git rev-parse --short HEAD)"

# if we're an automated build, use .GIT_COMMIT_AND_BRANCH as-is, else add -dirty
ifneq "$(origin BUILD_NUMBER)" "environment"
# not a CJE pipeline build
	ifneq "$(origin GITHUB_WORKSPACE)" "environment"
	# not a github build
	# assume non-pipeline build
	.GIT_COMMIT_AND_BRANCH := $(.GIT_COMMIT_AND_BRANCH)-dirty
	endif
endif

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

all: tools fmt lint reset build
rpm: rpm_package_source rpm_build_source rpm_build

help:
	@echo 'Usage: make <OPTIONS> ... <TARGETS>'
	@echo ''
	@echo 'Available targets are:'
	@echo ''
	@echo '    run                Run csi.'
	@echo '    help               Show this help screen.'
	@echo '    clean              Remove binaries, artifacts and releases.'
	@echo '    clean-artifacts    Remove build artifacts only.'
	@echo '    clean-releases     Remove releases only.'
	@echo '    tools              Install tools needed by the project.'
	@echo '    test               Run unit tests.'
	@echo '    vet                Run go vet.'
	@echo '    lint               Run golint.'
	@echo '    fmt                Run go fmt.'
	@echo '    tidy               Run go mod tidy.'
	@echo '    env                Display Go environment.'
	@echo '    build              Build project for current platform.'
	@echo '    doc                Start Go documentation server on port 8080.'
	@echo '    version            Display Go version.'
	@echo ''
	@echo 'Targets run by default are: fmt, lint, vet, and build.'
	@echo ''

print-%:
	@echo $* = $($*)

prepare:
	rm -rf $(BUILD_DIR)
	mkdir -p $(BUILD_DIR)/SPECS $(BUILD_DIR)/SOURCES
	cp $(SPEC_FILE) $(BUILD_DIR)/SPECS/

clean: clean-artifacts clean-releases
	go clean -i ./...
	rm -vf \
	  $(CURDIR)/build/results/coverage/* \
		$(CURDIR)/build/results/unittest/* \

clean-artifacts:
	rm -Rf artifacts/*

clean-releases:
	rm -Rf releases/*

clean-all: clean clean-artifacts

# Run tests
test: build
	go test ./cmd/... ./internal/... ./pkg/... -v -coverprofile coverage.out -covermode count

tools:
	go get -u github.com/axw/gocov/gocov
	go get -u github.com/AlekSi/gocov-xml
	go get -u golang.org/x/lint/golint
	go get -u github.com/t-yuki/gocover-cobertura
	go get -u github.com/jstemmer/go-junit-report
	go env -w GOPRIVATE=*.us.cray.com

vet: version
	go vet -v ./...

lint:
	golint -set_exit_status  ./...

fmt:
	go fmt ./...

env:
	@go env

# Run against the configured Kubernetes cluster in ~/.kube/configs
run: build
	go run ./main.go$(TARGET) $>

tidy:
	go mod tidy

reset:
	rm go.mod go.sum
	git checkout go.mod go.sum

build: fmt
	go build -o bin/csi -ldflags "\
	-X stash.us.cray.com/MTL/csi/pkg/version.gitVersion=${.GIT_VERSION} \
	-X stash.us.cray.com/MTL/csi/pkg/version.fsVersion=${.FS_VERSION} \
	-X stash.us.cray.com/MTL/csi/pkg/version.buildDate=${.BUILDTIME} \
	-X stash.us.cray.com/MTL/csi/pkg/version.sha1ver=${.GIT_COMMIT_AND_BRANCH}"
	bin/csi version

rpm_package_source:
	tar --transform 'flags=r;s,^,/$(SOURCE_NAME)/,' --exclude .git --exclude dist -cvjf $(SOURCE_PATH) .

rpm_build_source:
	BUILD_METADATA=$(BUILD_METADATA) rpmbuild -ts $(SOURCE_PATH) --define "_topdir $(BUILD_DIR)"

rpm_build:
	BUILD_METADATA=$(BUILD_METADATA) rpmbuild -ba $(SPEC_FILE) --define "_topdir $(BUILD_DIR)"

doc:
	godoc -http=:8080 -index

version:
	@go version

update-version: build
	@echo 'Version = ${CHANGELOG_VERSION}'
	echo ${CHANGELOG_VERSION} > .version
