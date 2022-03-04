FROM golang

WORKDIR $GOPATH/src/github.com/google/kne
COPY kne .

WORKDIR $GOPATH/src/github.com/openconfig/ondatra
COPY ondatra .
RUN go mod edit -replace=github.com/google/kne=$GOPATH/src/github.com/google/kne

RUN go mod tidy

RUN go build ./fakebind/main

RUN pwd
RUN echo $GOPATH

ENTRYPOINT ["go", "run", "fakebind/main/main.go"]
