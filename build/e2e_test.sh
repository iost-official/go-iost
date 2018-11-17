#!/bin/bash
NAME=$1

cd build/k8s/

echo "Generate iserver config"
go build -o iserver-config/genconfig iserver-config/genconfig.go
./iserver-config/genconfig -c $NAME -m 7 -s 2

echo "Delete test cluster $NAME in k8s"
kubectl delete -f iserver.yaml -n $NAME --ignore-not-found
kubectl delete pvc -l k8s-app=iserver -n $NAME
kubectl delete configmap iserver-config -n $NAME --ignore-not-found

echo "Create test cluster $NAME in k8s"
kubectl create configmap iserver-config --from-file=iserver-config -n $NAME
kubectl create -f iserver.yaml -n $NAME
kubectl create -f itest.yaml -n $NAME

echo "Wait 30s for creating cluster..."
sleep 30

echo "Start e2e test..."
kubectl exec -it itest -n $NAME -- ./itest run t_case

