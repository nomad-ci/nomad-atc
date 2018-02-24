FROM golang:alpine3.7 AS builder

RUN mkdir -p /go/src/github.com/nomad-ci

RUN apk add --no-cache git make

ADD . /go/src/github.com/nomad-ci/nomad-atc

RUN go get github.com/hashicorp/go-bindata/...
RUN cd /go/src/github.com/nomad-ci/nomad-atc && CGO_ENABLED=0 make build && cp bin/atc /tmp

FROM alpine:3.7

COPY --from=builder /tmp/atc /bin/atc

RUN apk add --no-cache ca-certificates

ENTRYPOINT ["/bin/atc"]
