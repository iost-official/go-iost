#!/bin/bash

NAME=$(git rev-parse --short HEAD)
if [ -n "$1" ]
then
    NAME=$1
fi

cd build/k8s/

echo "Delete test cluster $NAME in k8s"
kubectl delete -f iserver.yaml -n $NAME --ignore-not-found
kubectl delete -f itest.yaml -n $NAME --ignore-not-found
kubectl delete pvc -l k8s-app=iserver -n $NAME
kubectl delete configmap iserver-config -n $NAME --ignore-not-found
kubectl delete configmap itest-config -n $NAME --ignore-not-found
kubectl delete namespace $NAME --ignore-not-found

