FROM golang:1.26-bookworm AS builder

WORKDIR /src

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/fluxgate ./cmd/fluxgate

FROM gcr.io/distroless/static-debian12

WORKDIR /app
COPY --from=builder /out/fluxgate /app/fluxgate

ENV HTTP_ADDR=0.0.0.0:8080
ENV DB_PATH=/app/data/fluxgate.db
ENV LOG_DIR=/app/logs

EXPOSE 8080
ENTRYPOINT ["/app/fluxgate"]
