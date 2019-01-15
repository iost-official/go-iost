#!/bin/bash

readonly WITNESS_NUM=17

readonly WITNESS_NAME=(

)

readonly WITNESS_PUBKEY=(

)

readonly WITNESS_SECKEY=(

)

for i in {0..WITNESS_NUM-1}
do
    iwallet account --import ${WITNESS_NAME[i]} ${WITNESS_SECKEY[i]}
    iwallet --account ${WITNESS_NAME[i]} call "vote_producer.iost" "applyRegister" '["'${WITNESS_NAME[i]}'","'${WITNESS_PUBKEY[i]}'","location","url","",true]'
done
