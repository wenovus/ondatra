FROM golang

WORKDIR $GOPATH/src/github.com/openconfig/ondatra

ADD . .

RUN go mod tidy

RUN go build ./fakebind/main

RUN pwd
RUN echo $GOPATH

ENTRYPOINT ["go", "run", "fakebind/main/main.go", "-port", "1234", "-target", "fakedut"]

EXPOSE 1234
