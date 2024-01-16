FROM golang:1.21.5 AS builder

WORKDIR /build

COPY .. .

RUN go mod download

RUN GO111MODULE=on GOOS=linux GOARCH=amd64 go build -o service ./cmd/service

FROM scratch

COPY --from=builder /build/service /
COPY --from=builder /build/config/config.yaml /config/config.yaml
COPY --from=builder /build/data/quotes.txt /data/quotes.txt

EXPOSE 54345

ENTRYPOINT ["/service"]
