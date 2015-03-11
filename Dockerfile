FROM golang
MAINTAINER Aditya Mukerjee <dev@chimeracoder.net>

ADD . /go/src/github.com/ChimeraCoder/go-yo

RUN go install github.com/ChimeraCoder/go-yo

ENTRYPOINT ["/go/src/github.com/ChimeraCoder/go-yo/run.sh"]
