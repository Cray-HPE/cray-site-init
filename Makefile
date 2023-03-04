# MIT License
#
# (C) Copyright 2021-2022 Hewlett Packard Enterprise Development LP
#
# Permission is hereby granted, free of charge, to any person obtaining a
# copy of this software and associated documentation files (the "Software"),
# to deal in the Software without restriction, including without limitation
# the rights to use, copy, modify, merge, publish, distribute, sublicense,
# and/or sell copies of the Software, and to permit persons to whom the
# Software is furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included
# in all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
# THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
# OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
# ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
# OTHER DEALINGS IN THE SOFTWARE.
SHELL := /bin/bash -o pipefail
lc =$(subst A,a,$(subst B,b,$(subst C,c,$(subst D,d,$(subst E,e,$(subst F,f,$(subst G,g,$(subst H,h,$(subst I,i,$(subst J,j,$(subst K,k,$(subst L,l,$(subst M,m,$(subst N,n,$(subst O,o,$(subst P,p,$(subst Q,q,$(subst R,r,$(subst S,s,$(subst T,t,$(subst U,u,$(subst V,v,$(subst W,w,$(subst X,x,$(subst Y,y,$(subst Z,z,$1))))))))))))))))))))))))))

ifeq ($(NAME),)
export NAME := $(shell basename $(shell pwd))
endif

ifeq ($(ARCH),)
export ARCH := $(shell uname -m)
endif

ifeq ($(VERSION),)
export VERSION := $(shell git describe --tags | tr -s '-' '~' | tr -d '^v')
endif

# By default, if these are not set then set them to match the host.
ifeq ($(GOOS),)
OS := $(shell uname)
export GOOS := $(call lc,$(OS))
endif
ifeq ($(GOARCH),)
	ifeq "$(ARCH)" "aarch64"
		export GOARCH=arm64
	else ifeq "$(ARCH)" "x86_64"
		export GOARCH=amd64
	endif
endif

GO_FILES?=$$(find . -name '*.go' |grep -v vendor)
TAG?=latest

.GIT_COMMIT=$(shell git rev-parse HEAD)
.GIT_BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
.GIT_COMMIT_AND_BRANCH=$(.GIT_COMMIT)-$(subst /,-,$(.GIT_BRANCH))
.GIT_VERSION=$(shell git describe --tags 2>/dev/null || echo "$(.GIT_COMMIT)")
.BUILDTIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
CHANGELOG_VERSION_ORIG=$(grep -m1 \## CHANGELOG.MD | sed -e "s/\].*\$//" |sed -e "s/^.*\[//")
CHANGELOG_VERSION=$(shell grep -m1 \ \[[0-9]*.[0-9]*.[0-9]*\] CHANGELOG.MD | sed -e "s/\].*$$//" |sed -e "s/^.*\[//")
BUILD_DIR ?= $(PWD)/dist/rpmbuild
SPEC_FILE ?= ${NAME}.spec
SOURCE_NAME ?= ${NAME}-${VERSION}
SOURCE_PATH := ${BUILD_DIR}/SOURCES/${SOURCE_NAME}.tar.bz2
TEST_OUTPUT_DIR ?= $(CURDIR)/build/results

# if we're an automated build, use .GIT_COMMIT_AND_BRANCH as-is, else add '-dirty'
ifneq "$(origin BUILD_NUMBER)" "environment"
# not a Jenkins build
	ifneq "$(origin GITHUB_WORKSPACE)" "environment"
	# not a GitHub build
	# assume non-pipeline build
	.GIT_COMMIT_AND_BRANCH := $(.GIT_COMMIT_AND_BRANCH)-dirty
	endif
endif

.PHONY: \
	help \
	clean \
	tools \
	test \
	vet \
	lint \
	fmt \
	tidy \
	env \
	build \
	rpm \
	doc \
	version

all: bin/csi

rpm: rpm_prepare rpm_package_source rpm_build_source rpm_build

help:
	@echo 'Usage: make <OPTIONS> ... <TARGETS>'
	@echo ''
	@echo 'Available targets are:'
	@echo ''
	@echo '    help               Show this help screen.'
	@echo '    clean              Remove binaries, artifacts and releases.'
	@echo '    tools              Install tools needed by the project.'
	@echo '    test               Run unit tests.'
	@echo '    vet                Run go vet.'
	@echo '    lint               Run golint.'
	@echo '    fmt                Run go fmt.'
	@echo '    tidy               Run go mod tidy.'
	@echo '    env                Display Go environment.'
	@echo '    build              Build project for current platform.'
	@echo '    rpm                Build a YUM/SUSE RPM.'
	@echo '    doc                Start Go documentation server on port 8080.'
	@echo '    version            Display Go version.'
	@echo ''

clean:
	go clean -i ./...
	rm -vf \
	  $(CURDIR)/build/results/coverage/* \
	  $(CURDIR)/build/results/unittest/*
	rm -rf \
	  bin \
	  $(BUILD_DIR)

test: tools
	mkdir -pv $(TEST_OUTPUT_DIR)/unittest $(TEST_OUTPUT_DIR)/coverage
	go test ./cmd/... ./internal/... ./pkg/... -v -coverprofile $(TEST_OUTPUT_DIR)/coverage.out -covermode count | tee "$(TEST_OUTPUT_DIR)/testing.out"
	cat "$(TEST_OUTPUT_DIR)/testing.out" | go-junit-report | tee "$(TEST_OUTPUT_DIR)/unittest/testing.xml" | tee "$(TEST_OUTPUT_DIR)/unittest/testing.xml"
	gocover-cobertura < $(TEST_OUTPUT_DIR)/coverage.out > "$(TEST_OUTPUT_DIR)/coverage/coverage.xml"
	go tool cover -html=$(TEST_OUTPUT_DIR)/coverage.out -o "$(TEST_OUTPUT_DIR)/coverage/coverage.html"

tools:
	go install golang.org/x/lint/golint@latest
	go install github.com/t-yuki/gocover-cobertura@latest
	go install github.com/jstemmer/go-junit-report@latest

vet: version
	go vet -v ./...

lint: tools
	golint -set_exit_status ./cmd/...
	golint -set_exit_status ./internal/...
	golint -set_exit_status ./pkg/...

fmt:
	go fmt ./...

env:
	@go env

tidy:
	go mod tidy

bin/csi:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o bin/csi -ldflags "\
	-X github.com/Cray-HPE/cray-site-init/pkg/version.version=${.GIT_VERSION} \
	-X github.com/Cray-HPE/cray-site-init/pkg/version.buildDate=${.BUILDTIME} \
	-X github.com/Cray-HPE/cray-site-init/pkg/version.sha1ver=${.GIT_COMMIT_AND_BRANCH}"

rpm_prepare:
	rm -rf $(BUILD_DIR)
	mkdir -p $(BUILD_DIR)/SPECS $(BUILD_DIR)/SOURCES
	cp $(SPEC_FILE) $(BUILD_DIR)/SPECS/

rpm_package_source:
	tar --transform 'flags=r;s,^,/$(SOURCE_NAME)/,' --exclude .git --exclude dist -cvjf $(SOURCE_PATH) .

rpm_build_source:
	rpmbuild --nodeps --target $(ARCH) -ts $(SOURCE_PATH) --define "_topdir $(BUILD_DIR)"

rpm_build:
	rpmbuild --nodeps --target $(ARCH) -ba $(SPEC_FILE) --define "_topdir $(BUILD_DIR)"

doc:
	godoc -http=:8080 -index

version:
	@go version
