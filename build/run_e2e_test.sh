#!/bin/bash

NAME="devnet"

if [ -n "$1" ]
then
    NAME=$1
fi

echo "Start e2e test..."
kubectl exec -it itest -n $NAME -- ./itest -l debug run -c /etc/itest/itest.json t_case

