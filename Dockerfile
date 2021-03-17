FROM golang:1.15-alpine as builder

ARG VERSION

RUN apk add --no-cache git

RUN GO111MODULE=on go get github.com/korylprince/fileenv@v1.1.0

RUN git clone --branch "$VERSION" --single-branch --depth 1 \
    https://github.com/korylprince/snmp-tracker.git  /go/src/github.com/korylprince/snmp-tracker

RUN cd /go/src/github.com/korylprince/snmp-tracker && \
    go install -mod=vendor github.com/korylprince/snmp-tracker


FROM alpine:3.13

RUN apk add --no-cache ca-certificates

COPY --from=builder /go/bin/fileenv /
COPY --from=builder /go/bin/snmp-tracker /

CMD ["/fileenv", "/snmp-tracker"]
