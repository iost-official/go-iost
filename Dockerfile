FROM golang:1.10

RUN go get github.com/kardianos/govendor
# 复制代码
COPY . /go/src/github.com/iost-official/prototype/
# 安装依赖
RUN cd /go/src/github.com/iost-official/prototype/ && govendor sync -v
RUN go get github.com/golang/protobuf/proto github.com/iost-official/gopher-lua github.com/mitchellh/go-homedir
# RUN go get golang.org/x/net/context google.golang.org/grpc

RUN go build -o $GOPATH/bin/iwallet github.com/iost-official/prototype/iwallet
RUN go build  -o $GOPATH/bin/iserver github.com/iost-official/prototype/iserver

# 暴露端口
EXPOSE 30302
EXPOSE 30303

CMD /go/bin/iserver

# run: docker run -d -v /tmp:/tmp -p 30304:30304 iost-net
# build: docker build -t iost-net .