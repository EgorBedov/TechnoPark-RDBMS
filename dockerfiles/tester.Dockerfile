FROM golang:1.13-stretch

RUN go get -u -v github.com/bozaro/tech-db-forum
RUN go build github.com/bozaro/tech-db-forum
