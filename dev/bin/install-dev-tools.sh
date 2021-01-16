#!/bin/bash

set -eo pipefail

function run () {
    echo $1
    eval $1
}

function install_tools () {
    curl --silent --location \
         https://raw.githubusercontent.com/houseabsolute/ubi/master/bootstrap/bootstrap-ubi.sh |
        sh
    run "ubi --project golangci/golangci-lint --in ~/bin"
    run "go get golang.org/x/tools/cmd/goimports"
}

if [ "$1" == "-v" ]; then
    set -x
fi

install_tools

exit 0
