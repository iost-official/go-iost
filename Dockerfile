FROM golang:1.10

RUN go get github.com/kardianos/govendor
# copy source code
COPY . /go/src/github.com/iost-official/go-iost/
# install modules
RUN cd /go/src/github.com/iost-official/go-iost/ && govendor sync -v
RUN go get github.com/golang/protobuf/proto github.com/iost-official/gopher-lua github.com/mitchellh/go-homedir
RUN go get golang.org/x/net/context google.golang.org/grpc

# build iwallet cmd
RUN go build -o $GOPATH/bin/iwallet github.com/iost-official/go-iost/iwallet
# build iserver
RUN go build  -o $GOPATH/bin/iserver github.com/iost-official/go-iost/iserver

EXPOSE 30302
EXPOSE 30303

CMD /go/bin/iserver

# run: docker run -d -v /tmp:/tmp -p 30304:30304 iost-net
# build: docker build -t iost-net .
