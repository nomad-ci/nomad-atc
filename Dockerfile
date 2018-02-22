FROM golang:alpine3.7 AS builder

RUN mkdir -p /go/src/github.com/nomad-ci

ADD . /go/src/github.com/nomad-ci/nomad-atc

RUN go build -o /tmp/atc github.com/nomad-ci/nomad-atc/cmd/atc

FROM alpine:3.7

COPY --from=builder /tmp/atc /bin/atc

RUN apk add --no-cache ca-certificates

ENTRYPOINT ["/bin/atc"]
