#!/bin/bash

set -e

if [[ "$1" != "binary-only" ]]; then
    GO_VERSION="1.16.3"
    INSTALLED_GO_VERSION=$(go version | awk '{print $3}')

    if [[ "go${GO_VERSION}" !=  $INSTALLED_GO_VERSION ]]; then
        echo "Upgrading go from version ${INSTALLED_GO_VERSION} to ${GO_VERSION}"
        go get golang.org/dl/go$GO_VERSION || true
        $GOPATH/bin/go$GO_VERSION download || true
        GO_EXEC=$(which go)
        rm -f $GO_EXEC
        cp $GOPATH/bin/go$GO_VERSION $GO_EXEC
    fi

    mkdir -p $GOPATH/bin
    mkdir -p $GOPATH/src
    mkdir -p $GOPATH/pkg
fi