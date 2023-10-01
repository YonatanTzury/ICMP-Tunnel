#!/bin/bash

SCRIPT_PATH="$(dirname -- "${BASH_SOURCE[0]}")"

(ls $SCRIPT_PATH/build && rm $SCRIPT_PATH/build/*) 2> /dev/null
(ls $SCRIPT_PATH/build || mkdir $SCRIPT_PATH/build) 2> /dev/null

go build -o $SCRIPT_PATH/build/client $SCRIPT_PATH/cmd/client/client.go
go build -o $SCRIPT_PATH/build/server $SCRIPT_PATH/cmd/server/server.go
