#!/bin/bash

NAME=$(git rev-parse --short HEAD)
if [ -n "$1" ]
then
    NAME=$1
fi

echo "Start e2e test..."
kubectl exec -it itest -n $NAME -- ./itest run -c /etc/itest/itest.json t_case

