#!/bin/bash

set -e

ERR=$( go mod tidy -v 2>&1 )
if [[ "$ERR" =~ "unused" ]]; then
    echo $ERR
    exit 1
fi

exit 0
