#!/bin/bash
set -e -u

if [ "$(go version 2>/dev/null)" != "go version go1.8.3 linux/amd64" ]
then
	echo "go version mismatch! expected 1.8.3" 1>&2
	go version 1>&2
	exit 1
fi

export GOPATH="$(pwd)"

go build src/farad/main/farad.go
# go build src/faradayd/main/faradayd.go
