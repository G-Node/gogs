FROM golang:1.8

WORKDIR /go/src/github.com/G-Node/gin-dex
COPY . .

RUN  go get ./...
RUN go build

EXPOSE 8099
VOLUME /repos

ENTRYPOINT  ./gin-dex --debug
