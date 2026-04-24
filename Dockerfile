FROM golang:1.24-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o dockhand_exporter .

# ── minimal runtime image ─────────────────────────────────────────────────────
FROM scratch

COPY --from=builder /build/dockhand_exporter /dockhand_exporter

EXPOSE 9090

ENTRYPOINT ["/dockhand_exporter"]
