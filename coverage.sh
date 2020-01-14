#!/usr/bin/env bash

set -e

echo "mode: atomic" > coverage.tx

PKG_LIST=$(go list ./... | grep -v /vendor/)
for package in ${PKG_LIST}; do
    go test -covermode=atomic -coverprofile "coverage/${package##*/}.out" "$package" ;
    if [ -f "coverage/${package##*/}.out" ]; then
        echo "$(pwd)"
        cat "coverage/${package##*/}.out" | grep -v "mode: " >> coverage.tx
        rm "coverage/${package##*/}.out"
    fi
done