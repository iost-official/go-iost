#!/bin/bash

NAME=$(git rev-parse --short HEAD)
if [ -n "$1" ]
then
    NAME=$1
fi

cd build/k8s/

echo "Generate iserver config"
cd iserver-config/
go run genconfig.go -c $NAME -m 7 -s 2
cd -

echo "Create test cluster $NAME in k8s"
kubectl create namespace $NAME
kubectl create configmap iserver-config --from-file=iserver-config -n $NAME
kubectl create -f iserver.yaml -n $NAME
kubectl create -f itest.yaml -n $NAME

