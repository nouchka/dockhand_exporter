FROM golang:1.24-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o dockhand_exporter . && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o healthcheck ./healthcheck

# ── minimal runtime image ─────────────────────────────────────────────────────
FROM scratch

COPY --from=builder /build/dockhand_exporter /dockhand_exporter
COPY --from=builder /build/healthcheck /healthcheck

EXPOSE 9090

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD ["/healthcheck"]

ENTRYPOINT ["/dockhand_exporter"]
