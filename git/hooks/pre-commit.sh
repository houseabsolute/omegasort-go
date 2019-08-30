#!/bin/bash

status=0

go generate ./...
if (( $? != 0 )); then
    status+=1
fi

./bin/precious lint -s
if (( $? != 0 )); then
    status+=2
fi

exit $status
