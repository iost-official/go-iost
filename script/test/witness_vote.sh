#!/bin/bash


readonly VOTE_ACCOUNT="admin"
readonly VOTE_ACCOUNT_SECKEY="2yquS3ySrGWPEKywCPzX4RTJugqRh7kJSo5aehsLYPEWkUxBWA39oMrZ7ZxuM4fgyXYs2cPwh5n8aNNpH5x2VyK1"

readonly WITNESS_NUM=17

readonly WITNESS_NAME=(

)

iwallet account --import ${VOTE_ACCOUNT} ${VOTE_ACCOUNT_SECKEY}
for i in {0..IOST_WITNESS_NUM-1}
do
    iwallet --account ${VOTE_ACCOUNT} call "vote_producer.iost" "vote" '["'${VOTE_ACCOUNT}'", "'${IOST_WITNESS_NAME[i]}'", "3000000"]'
done