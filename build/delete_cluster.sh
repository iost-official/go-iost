#!/bin/bash

NAME="devnet"
COMMIT=$(git rev-parse --short HEAD)

if [ -n "$1" ]
then
    NAME=$1
fi

cd build/k8s/

echo "Delete test cluster $NAME in k8s"
cat iserver.yaml | sed 's/\$COMMIT'"/$COMMIT/g" | kubectl delete -f - -n $NAME --ignore-not-found
cat itest.yaml | sed 's/\$COMMIT'"/$COMMIT/g" | kubectl delete -f - -n $NAME --ignore-not-found
kubectl delete pvc -l k8s-app=iserver -n $NAME
kubectl delete configmap iserver-config -n $NAME --ignore-not-found
kubectl delete configmap itest-config -n $NAME --ignore-not-found

