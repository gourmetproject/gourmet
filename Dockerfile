FROM golang:1.13

RUN apt-get update && apt-get install -y libpcap-dev

RUN go get -u github.com/gourmetproject/gourmet
RUN go get gopkg.in/yaml.v2

WORKDIR $GOROOT/github.com/gourmetproject/gourmet

COPY config.yml .
RUN make build

ENTRYPOINT ["./bin/gourmet"]
