#!/bin/bash

readonly GRPC_URL="127.0.0.1:30002"
readonly WITNESS_NUM=1

readonly WITNESS_NAME=(
producer000
)

readonly WITNESS_SECKEY=(
1rANSfcRzr4HkhbUFZ7L1Zp69JZZHiDDq5v7dNSbbEqeU4jxy3fszV4HGiaLQEyqVpS1dKT9g7zCVRxBVzuiUzB
)

for (( i = 0; i < ${WITNESS_NUM}; i++ ))
do
    iwallet -s ${GRPC_URL} account --import ${WITNESS_NAME[i]} ${WITNESS_SECKEY[i]}
    iwallet -s ${GRPC_URL} --account ${WITNESS_NAME[i]} call 'vote_producer.iost' 'logInProducer' '["'${WITNESS_NAME[i]}'"]'
done
