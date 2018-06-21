
# Test

## Build Image
根据当前代码编译生成docker image，注意：build下的binary会变成centos下的
```
make image
```

## Set ENV
设置一些关键环境变量
```
export DOCKER_IMAGE="iost-node:1.0.0-$(git rev-parse --short HEAD)"
export PROJECT="$GOPATH/src/github.com/iost-official/prototype"
export LOCAL_IP="$(ipconfig getifaddr en0)"
```

## Run prometheus server
启动本地prometheus服务，方便查看metrics
```
mkdir -p test/data/prometheus
cp test/template/prometheus.yml test/data/prometheus/
sed -i '.bak' "s/{{LOCAL_IP}}/${LOCAL_IP}/g" test/data/prometheus/prometheus.yml
docker run -d -p 9090:9090 --name prometheus \
       -v $PROJECT/test/data/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml \
       prom/prometheus
```

## Run register server
运行注册服务，注意mode需要为private
```
mkdir -p test/data/register
docker run -d -p 30304:30304 --name iost_register \
       -v $PROJECT/test/data/register:/workdir/data \
       $DOCKER_IMAGE ./register --mode private
```

## Run 3 iserver
运行3个iserver服务，加上`--cpuprofile`参数后，在进程正常退出时会生成cpu的profile文件
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
       $DOCKER_IMAGE ./start.sh --cpuprofile /var/lib/iserver/cpu.prof
docker run -d -p 30312:30312 -p 30313:30313 -p 8081:8080 --name iost_iserver1 \
       -v $PROJECT/test/data/iserver1:/var/lib/iserver \
       $DOCKER_IMAGE ./start.sh --cpuprofile /var/lib/iserver/cpu.prof
docker run -d -p 30322:30322 -p 30323:30323 -p 8082:8080 --name iost_iserver2 \
       -v $PROJECT/test/data/iserver2:/var/lib/iserver \
       $DOCKER_IMAGE ./start.sh --cpuprofile /var/lib/iserver/cpu.prof
```

## Browser prometheus
通过本地prometheus查看关键metrics
[http://127.0.0.1:9090](http://127.0.0.1:9090/graph?g0.range_input=1h&g0.expr=rate(generated_block_count%5B1m%5D)&g0.tab=0&g1.range_input=1h&g1.expr=rate(received_block_count%5B1m%5D)&g1.tab=0&g2.range_input=1h&g2.expr=rate(received_transaction_count%5B1m%5D)&g2.tab=0&g3.range_input=1h&g3.expr=confirmed_blockchain_length&g3.tab=0)

## Build txsender
因为`make image`时会使用centos环境编译所有cmd，所以想在本机执行，需要重新编译，文件会在build目录下：
```
make txsender
```

## Run txsender
执行一段时间txsender，产生负载，然后关闭txsender
```
./build/txsender -cluster local
```

## Exit all server
正常退出所有服务
```
docker stop iost_iserver0
docker stop iost_iserver1
docker stop iost_iserver2
docker stop iost_register
docker stop prometheus
```

## View cpu.prof
使用pprof工具查看性能检测结果，通过函数调用图，火焰图等进行性能调优（因为binary需要使用centos下的，所以需要重新用`make image`编译txsender）
```
make image
pprof -http "127.0.0.1:12345" build/iserver test/data/iserver0/cpu.prof 
```

## Remove all server
清理所有服务docker container，方便之后重新测试
```
docker rm iost_iserver0
docker rm iost_iserver1
docker rm iost_iserver2
docker rm iost_register
docker rm prometheus
```
