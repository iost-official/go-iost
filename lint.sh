#! /bin/bash

gometalinter --config $GOPATH/src/github.com/iost-official/prototype/gometalinter.json $GOPATH/src/github.com/iost-official/prototype/... |grep -v 'exported method\|function'
echo "ok"