FROM golang:1.6
MAINTAINER Octoblu, Inc. <docker@octoblu.com>

WORKDIR /go/src/github.com/octoblu/meshblu-connector-ignition
COPY . /go/src/github.com/octoblu/meshblu-connector-ignition

RUN env CGO_ENABLED=0 go build -o meshblu-connector-ignition -a -ldflags '-s' .

CMD ["./meshblu-connector-ignition"]
