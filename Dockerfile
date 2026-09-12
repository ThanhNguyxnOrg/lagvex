FROM golang:1.22-alpine AS builder

WORKDIR /src
COPY go.mod ./
COPY pkg/ ./pkg/
COPY cmd/ ./cmd/

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /bin/lagvex-relay ./cmd/lagvex-relay

FROM alpine:3.20

RUN apk add --no-cache iptables iproute2 bash
COPY --from=builder /bin/lagvex-relay /usr/local/bin/lagvex-relay

ENV LAGVEX_LISTEN=":51820" \
    LAGVEX_TUN="lagvex0" \
    LAGVEX_SUBNET="10.88.0.0/24" \
    LAGVEX_MTU="1400"

EXPOSE 51820/udp

ENTRYPOINT ["/usr/local/bin/lagvex-relay"]
