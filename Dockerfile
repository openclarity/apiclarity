## Build Frontend
FROM --platform=$BUILDPLATFORM node:14-slim AS site-build

WORKDIR /app/ui-build

# Cache optimization: Avoid npm install unless package.json changed
COPY ui/package-lock.json ui/package.json ./
RUN npm ci

COPY ui .
RUN npm run build

## Build Backend
# Cross-compilation tools
FROM --platform=$BUILDPLATFORM tonistiigi/xx AS xx

FROM --platform=$BUILDPLATFORM golang:1.16.6-alpine AS builder

# Copy cross-compilation tools
COPY --from=xx / /

WORKDIR /build
COPY api ./api
COPY plugins/api ./plugins/api

# Cache optimization: Avoid go module downloads unless go.mod/go.sum changed
WORKDIR /build/backend
COPY backend/go.* ./
RUN go mod download

ARG BUILD_TIMESTAMP COMMIT_HASH VERSION TARGETOS TARGETARCH

# Copy and build backend code
COPY backend .
ARG TARGETPLATFORM
ENV CGO_ENABLED=0
RUN xx-go build -ldflags="-s -w \
     -X 'github.com/apiclarity/apiclarity/backend/pkg/version.Version=${VERSION}' \
     -X 'github.com/apiclarity/apiclarity/backend/pkg/version.CommitHash=${COMMIT_HASH}' \
     -X 'github.com/apiclarity/apiclarity/backend/pkg/version.BuildTimestamp=${BUILD_TIMESTAMP}'" -o backend ./cmd/backend/main.go

FROM alpine:3.14

WORKDIR /app

COPY --from=builder ["/build/backend/pkg/test/trace_files/", "trace_files"]
COPY --from=builder ["/build/backend/pkg/test/provided_spec/", "provided_spec"]
COPY --from=builder ["/build/backend/pkg/test/diff_trace_files/", "diff_trace_files"]
COPY --from=builder ["/build/backend/backend", "./backend"]
COPY --from=site-build ["/app/ui-build/build", "site"]
COPY dist dist

ENTRYPOINT ["/app/backend"]
