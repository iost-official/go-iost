# Run yourself private network by docker
## Build docker image
Generate a docker image based on the current code. Note that the binary file in the build directory will be the centos system binary file.
```
make image
```

## Set environment variable
Set some required environment variables.
### Linux
```
export DOCKER_IMAGE="iost-node:1.0.3-$(git rev-parse --short HEAD)"
export PROJECT=`pwd`
export LOCAL_IP="hostname -i"
```
### Mac OS X
```
export DOCKER_IMAGE="iost-node:1.0.3-$(git rev-parse --short HEAD)"
export PROJECT=`pwd`
export LOCAL_IP="$(ipconfig getifaddr en0)"
```

## Run register server
Run the register server, note that mode needs to be private.
```
mkdir -p test/data/register
docker run -d -p 30304:30304 --name iost_register \
       -v $PROJECT/test/data/register:/workdir/data \
       $DOCKER_IMAGE ./register --mode private
```

## Run three iservers
First create three iserver working directories, then generate three iserver configuration files, and finally run the servers.
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

## Check the logs
You can view the iserver0 logs in the `test/data/iserver0/logs/current.log` file，for example：
```
tail -f test/data/iserver0/logs/current.log 
tail -f test/data/iserver1/logs/current.log 
tail -f test/data/iserver2/logs/current.log 
```

## Exit all server
Exit all server normally.
```
docker stop iost_iserver0
docker stop iost_iserver1
docker stop iost_iserver2
docker stop iost_register
```

## Remove all server
Clean up all server.
```
docker rm -f iost_iserver0
docker rm -f iost_iserver1
docker rm -f iost_iserver2
docker rm -f iost_register
```
