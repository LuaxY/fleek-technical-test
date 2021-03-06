FROM golang:1.14-alpine AS builder
WORKDIR /app
COPY . .
RUN go get -u ./...
RUN CGO_ENABLED=0 go test ./...
RUN go build -o server ./cmd/server

FROM alpine AS server
RUN mkdir -p /data/unencrypted
RUN mkdir -p /data/encrypted
COPY --from=builder /app/server /server
COPY web /web
ENTRYPOINT ["/server", "-src", "/data/unencrypted", "-dst", "/data/encrypted"]