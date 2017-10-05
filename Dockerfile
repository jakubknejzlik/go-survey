FROM golang:alpine

RUN apk update && \
    apk add alpine-sdk

COPY . $GOPATH/src/github.com/jakubknejzlik/go-survey

WORKDIR $GOPATH/src/github.com/jakubknejzlik/go-survey

RUN make install

ENTRYPOINT ["go-survey"]
