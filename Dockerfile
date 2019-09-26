FROM golang:1.13

RUN apt-get update && \
    apt-get -y -q install wget lsb-release gnupg && \
    wget -q http://apt.ntop.org/buster/all/apt-ntop.deb && dpkg -i apt-ntop.deb && \
    apt-get clean all && \
    apt-get update && \
    apt-get -y install pfring libpcap-dev

RUN go get -u github.com/gourmetproject/gourmet
RUN go get gopkg.in/yaml.v2

COPY config.yml .

RUN go build /go/src/github.com/gourmetproject/gourmet/cmd/main.go

ENTRYPOINT ["./main"]
