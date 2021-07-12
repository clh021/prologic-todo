#!/bin/bash
set -euo pipefail

export GOPATH=$HOME/go

[ -d $GOPATH/src/git.mills.io/prologic/ ] || mkdir -p $GOPATH/src/git.mills.io/prologic/
[ -L $GOPATH/src/git.mills.io/prologic/todo ] || \
	ln -s /opt/app $GOPATH/src/git.mills.io/prologic/todo

cd /opt/app
go get -v -d ./...
make build
exit 0
