#!/bin/bash


readonly GRPC_URL="127.0.0.1:30002"
readonly CREATOR_ACCOUNT="admin"
readonly CREATOR_ACCOUNT_SECKEY="2yquS3ySrGWPEKywCPzX4RTJugqRh7kJSo5aehsLYPEWkUxBWA39oMrZ7ZxuM4fgyXYs2cPwh5n8aNNpH5x2VyK1"

readonly WITNESS_NUM=1

readonly WITNESS_NAME=(
producer000
)

readonly WITNESS_PUBKEY=(
6sNQa7PV2SFzqCBtQUcQYJGGoU7XaB6R4xuCQVXNZe6b
)

iwallet account --import ${CREATOR_ACCOUNT} ${CREATOR_ACCOUNT_SECKEY}
for (( i = 0; i < ${WITNESS_NUM}; i++ ))
do
    iwallet -s ${GRPC_URL} --account ${CREATOR_ACCOUNT} --amount_limit "ram:1024|iost:10" account --create ${WITNESS_NAME[i]} --initial_balance 0 --initial_gas_pledge 10 --initial_ram 1024 --owner ${WITNESS_PUBKEY[i]} --active ${WITNESS_PUBKEY[i]}
done
