FROM golang:1.26-bookworm AS builder

WORKDIR /src
ARG GOPROXY=https://goproxy.cn,https://proxy.golang.org,direct
ENV GOPROXY=${GOPROXY}

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/fluxgate ./cmd/fluxgate

FROM scratch

WORKDIR /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /out/fluxgate /app/fluxgate

ENV HTTP_ADDR=0.0.0.0:8080
ENV DB_PATH=/app/data/fluxgate.db
ENV LOG_DIR=/app/logs

EXPOSE 8080
ENTRYPOINT ["/app/fluxgate"]
