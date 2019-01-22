#!/bin/bash

NAME="devnet"
COMMIT=$(git rev-parse --short HEAD)

if [ -n "$1" ]
then
    NAME=$1
fi

cd build/k8s/

echo "Generate iserver config"
cd iserver-config/
go run genconfig.go -c $NAME -m 3 -s 1
cp -r ${GOPATH}/src/github.com/iost-official/go-iost/config/genesis/contract .
cd -

echo "Generate itest config"
cd itest-config/
export DYLD_LIBRARY_PATH=${GOPATH}/src/github.com/iost-official/go-iost/vm/v8vm/v8/libv8/_darwin_amd64
go run genconfig.go -c $NAME -s "3"
cd -

echo "Create test cluster $NAME in k8s"
kubectl create configmap iserver-config --from-file=iserver-config -n $NAME
kubectl create configmap iserver-contract --from-file=iserver-config/contract -n $NAME
cat iserver.yaml | sed 's/\$COMMIT'"/$COMMIT/g" | kubectl create -f - -n $NAME
kubectl create configmap itest-config --from-file=itest-config -n $NAME
cat itest.yaml | sed 's/\$COMMIT'"/$COMMIT/g" | kubectl create -f - -n $NAME

