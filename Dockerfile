FROM golang:1.19 AS builder
WORKDIR /go/src/github.com/chrisohaver/dnsdrone
COPY main.go go.mod go.sum .
RUN go build -o dnsdrone .

FROM debian:stable-slim
WORKDIR /go/src/github.com/chrisohaver/dnsdrone
COPY --from=builder /go/src/github.com/chrisohaver/dnsdrone .
ENTRYPOINT ["./dnsdrone"]

