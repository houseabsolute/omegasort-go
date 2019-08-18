#!/bin/bash

status=0

golangci-lint run \
    -c ./.golangci-lint-main.yml \
    --skip-dirs internal/.+

status+=$?

golangci-lint run \
    -c ./.golangci-lint-internal.yml \
    --skip-dirs internal/guesswidth \
    --skip-dirs internal/posixpath \
    --skip-dirs internal/winpath \
    ./internal/...

status+=$?

exit $status
