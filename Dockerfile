FROM golang:1.13 AS builder
WORKDIR /go/src/github.com/chrisohaver/dnsdrone
COPY main.go .
RUN go get -d -v && go build -o dnsdrone .

FROM debian:stable-slim
WORKDIR /go/src/github.com/chrisohaver/dnsdrone
COPY --from=builder /go/src/github.com/chrisohaver/dnsdrone .
ENTRYPOINT ["./dnsdrone"]