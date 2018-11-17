# iserver

## Generate iserver config
```
cd iserver-config
go run genconfig.go devnet
```

## Create iserver
```
kubectl create configmap iserver-config --from-file=iserver-config -n devnet
kubectl create -f iserver.yaml -n devnet
```

## Delete iserver
```
kubectl delete -f iserver.yaml -n devnet
kubectl delete pvc -l k8s-app=iserver -n devnet
kubectl delete configmap iserver-config -n devnet
```
