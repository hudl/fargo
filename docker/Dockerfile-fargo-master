FROM golang:1.6

ENV GOROOT=/usr/local/go

RUN go get github.com/tools/godep
RUN go get github.com/hudl/fargo

WORKDIR /go/src/github.com/hudl/fargo/
RUN godep restore