#! /bin/bash

rm /workdir/iserver.yml
cp $GOPATH/src/github.com/iost-official/prototype/iserver/$1 /workdir/iserver.yml && echo "ok"