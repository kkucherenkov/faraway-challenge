FROM golang:1.21.5 AS builder

WORKDIR /build

COPY .. .

RUN go mod download

RUN GO111MODULE=on GOOS=linux GOARCH=amd64 go build -o client ./cmd/client

FROM scratch

COPY --from=builder /build/client /
COPY --from=builder /build/config/config.yaml /config/config.yaml

ENTRYPOINT ["/client"]
