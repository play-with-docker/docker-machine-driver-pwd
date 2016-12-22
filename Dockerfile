FROM golang:1.7

RUN go get github.com/franela/docker-machine-driver-pwd

COPY . $GOPATH/src/github.com/franela/docker-machine-driver-pwd

RUN go install github.com/franela/docker-machine-driver-pwd

CMD ["docker-machine-driver-pwd"]
