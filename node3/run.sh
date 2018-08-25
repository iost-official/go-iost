cd ..
make
cd -
rm -r leveldb
rm -r StatePoolDB
rm -r logs
rm priv.key
rm routing.table
../target/iserver -f iserver.yaml
