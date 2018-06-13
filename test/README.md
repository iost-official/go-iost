
# Test

## Build Image
```
make image
```

## Set ENV
```
export DOCKER_IMAGE="iost-node:1.0.0-$(git rev-parse --short HEAD)"
export PROJECT="$GOPATH/src/github.com/iost-official/prototype"
export LOCAL_IP="$(ipconfig getifaddr en0)"
```

## Run prometheus server
```
mkdir -p test/data/prometheus
cp test/template/prometheus.yml test/data/prometheus/
sed -i '.bak' "s/{{LOCAL_IP}}/${LOCAL_IP}/g" test/data/prometheus/prometheus.yml
docker run -d -p 9090:9090 --name prometheus \
       -v $PROJECT/test/data/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml \
       prom/prometheus
```

## Run register server
```
mkdir -p test/data/register
docker run -d -p 30304:30304 --name iost_register \
       -v $PROJECT/test/data/register:/workdir/data \
       $DOCKER_IMAGE ./register --mode private
```

## Run 3 iserver
```
mkdir -p test/data/iserver0
mkdir -p test/data/iserver1
mkdir -p test/data/iserver2
cp test/template/iserver0.yml test/data/iserver0/iserver.yml
cp test/template/iserver1.yml test/data/iserver1/iserver.yml
cp test/template/iserver2.yml test/data/iserver2/iserver.yml
sed -i '.bak' "s/{{LOCAL_IP}}/${LOCAL_IP}/g" test/data/iserver0/iserver.yml
sed -i '.bak' "s/{{LOCAL_IP}}/${LOCAL_IP}/g" test/data/iserver1/iserver.yml
sed -i '.bak' "s/{{LOCAL_IP}}/${LOCAL_IP}/g" test/data/iserver2/iserver.yml

docker run -d -p 30302:30302 -p 30303:30303 -p 8080:8080 --name iost_iserver0 \
       -v $PROJECT/test/data/iserver0:/var/lib/iserver \
       $DOCKER_IMAGE ./start.sh
docker run -d -p 30312:30312 -p 30313:30313 -p 8081:8080 --name iost_iserver1 \
       -v $PROJECT/test/data/iserver1:/var/lib/iserver \
       $DOCKER_IMAGE ./start.sh
docker run -d -p 30322:30322 -p 30323:30323 -p 8082:8080 --name iost_iserver2 \
       -v $PROJECT/test/data/iserver2:/var/lib/iserver \
       $DOCKER_IMAGE ./start.sh
```

## Browser prometheus
[http://127.0.0.1:9090](http://127.0.0.1:9090)

## Exit all server
```
docker rm -f prometheus
docker rm -f iost_register
docker rm -f iost_iserver0
docker rm -f iost_iserver1
docker rm -f iost_iserver2
```
