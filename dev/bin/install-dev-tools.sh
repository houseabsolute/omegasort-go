#!/bin/bash

set -eo pipefail

function run () {
    echo $1
    eval $1
}

function install_tools () {
    run "./dev/bin/download-golangci-lint.sh v1.17.1"
    run "./dev/bin/download-precious.sh"
    run "go get golang.org/x/tools/cmd/goimports"
}

if [ "$1" == "-v" ]; then
    set -x
fi

install_tools

exit 0
