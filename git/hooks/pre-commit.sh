#!/bin/bash

status=0

go generate ./...
if (( $? != 0 )); then
    status+=1
fi

./dev/bin/run-golangci-lint.sh
if (( $? != 0 )); then
    status+=2
fi

exit $status
