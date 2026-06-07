# ── Stage 1: Build ──────────────────────────────────────────────────────────
# golang:1.22-alpine is used (not debian) to keep the builder layer small and
# avoid pulling in glibc.
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Cache dependency downloads separately from source changes.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Run tests before building - the final binary is only produced if tests pass.
RUN go test ./...

# -ldflags="-s -w" strips the symbol table and DWARF debug info, reducing
# binary size by ~30%. CGO_ENABLED=0 ensures a fully static binary.
# CGO is disabled so the binary is statically linked
# and compatible with the distroless/static base image.
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o csv-processor .

# ── Stage 2: Run ─────────────────────────────────────────────────────────────
# alpine:3.19 instead of distroless: the Nginx sidecar pattern requires an
# init container that runs `cp -r /static/. /shared/static/` to populate the
# shared emptyDir volume. Distroless has no shell, making the same image
# unusable as both init container and app container.
#
# Using Alpine means one image, one CI build, one tag to manage. The shell's
# presence is compensated by Kubernetes securityContext on every container:
#   - runAsNonRoot: true  (uid 65534, nobody)
#   - readOnlyRootFilesystem: true
#   - allowPrivilegeEscalation: false
#   - capabilities: drop: [ALL]
# See decisions.md - "Why Alpine over distroless for the runtime image."
FROM alpine:3.19

COPY --from=builder /app/csv-processor /csv-processor
COPY --from=builder /app/templates /templates
COPY --from=builder /app/static /static

# nobody (uid 65534) is built into Alpine — no adduser needed.
USER nobody

EXPOSE 8080

ENTRYPOINT ["/csv-processor"]
