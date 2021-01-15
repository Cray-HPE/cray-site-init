#!/bin/bash

export GOPATH="$HOME/go"
export PATH="$PATH:$GOPATH/bin"

export TEST_OUTPUT_DIR="$PWD/build/results"
mkdir -pv "$TEST_OUTPUT_DIR/coverage" "$TEST_OUTPUT_DIR/unittest"
make clean && make test | go-junit-report -set-exit-code > "$TEST_OUTPUT_DIR/unittest/testing.xml"
gocover-cobertura < cover.out > "$TEST_OUTPUT_DIR/coverage/coverage.xml"
