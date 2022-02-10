FROM golang:1.17 as golang

ADD . $GOPATH/src/logstash_exporter/

WORKDIR $GOPATH/src/logstash_exporter/
RUN go mod download && go build .

FROM busybox:1.27.2-glibc
COPY --from=golang /go/src/logstash_exporter/logstash_exporter /
EXPOSE 9198
ENTRYPOINT ["/logstash_exporter"]  
