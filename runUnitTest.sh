#!/bin/bash

export GOPATH="$HOME/go"
export PATH="$PATH:$GOPATH/bin"

export TEST_OUTPUT_DIR="$PWD/build/results"
mkdir -pv "$TEST_OUTPUT_DIR/coverage" "$TEST_OUTPUT_DIR/unittest"
make clean && make test | go-junit-report -set-exit-code | tee "$TEST_OUTPUT_DIR/unittest/testing.xml"
gocover-cobertura < coverage.out > "$TEST_OUTPUT_DIR/coverage/coverage.xml"
go tool cover -html=coverage.out -o "$TEST_OUTPUT_DIR/coverage/coverage.html"
