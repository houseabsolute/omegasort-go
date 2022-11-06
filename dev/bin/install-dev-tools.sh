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
    run "ubi --project houseabsolute/precious --in ~/bin"
    run "ubi --project golangci/golangci-lint --in ~/bin"
    # If we run this in the checkout dir it can mess with out go.mod and
    # go.sum.
    pushd /tmp
    run "go install golang.org/x/tools/cmd/goimports@latest"
    popd
}

if [ "$1" == "-v" ]; then
    set -x
fi

install_tools

exit 0
