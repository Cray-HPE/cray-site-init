#!/bin/bash

export GOPATH="$HOME/go"
export PATH="$PATH:$GOPATH/bin"

make clean && make test

if [[ -f coverage.out ]]; then
  export TEST_OUTPUT_DIR="$PWD/build/results/unittest/"
  mkdir -pv "$TEST_OUTPUT_DIR"
  mv coverage.out "$TEST_OUTPUT_DIR"
fi
