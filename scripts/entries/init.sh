#! /bin/bash

rm $GOPATH/src/github.com/iost-official/prototype/iserver/iserver.yml
cp $GOPATH/src/github.com/iost-official/prototype/iserver/$1 $GOPATH/src/github.com/iost-official/prototype/iserver/iserver.yml && echo "ok"