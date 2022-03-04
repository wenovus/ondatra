FROM golang

WORKDIR $GOPATH/src/github.com/
COPY openconfig/lemming openconfig/lemming
COPY wenovus/kne google/kne
COPY wenovus/ondatra openconfig/ondatra

WORKDIR $GOPATH/src/github.com/openconfig/ondatra
RUN go mod edit -replace=github.com/openconfig/lemming=$GOPATH/src/github.com/openconfig/lemming
RUN go mod edit -replace=github.com/google/kne=$GOPATH/src/github.com/google/kne

RUN go mod tidy

RUN pwd
RUN echo $GOPATH

WORKDIR $GOPATH/src/github.com/openconfig/lemming
RUN go build ./gnmi/main
ENTRYPOINT ["go", "run", "gnmi/main/main.go"]
