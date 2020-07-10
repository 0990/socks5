FROM golang:1.13 AS builder
COPY . socks5
WORKDIR socks5

RUN go env -w GOPROXY="https://goproxy.cn,direct"
RUN CGO_ENABLED=0 go build -o /bin/ss5 ./cmd/server/main.go

FROM scratch
COPY --from=builder /bin/ss5 .
CMD ["./ss5"]