FROM node:14-slim AS site-build

WORKDIR /app/ui-build

COPY ui .
RUN npm i
RUN npm run build


FROM golang:1.16.6-alpine AS builder

RUN apk add --update --no-cache gcc g++

WORKDIR /build
COPY api ./api

WORKDIR /build/backend
COPY backend/go.* ./
RUN go mod download

# Copy and build backend code
COPY backend .
RUN go build -o backend ./cmd/backend/main.go


FROM alpine:3.14.0

WORKDIR /app

COPY --from=builder ["/build/backend/pkg/test/trace_files/", "trace_files"]
COPY --from=builder ["/build/backend/pkg/test/provided_spec/", "provided_spec"]
COPY --from=builder ["/build/backend/pkg/test/diff_trace_files/", "diff_trace_files"]
COPY --from=builder ["/build/backend/backend", "./backend"]
COPY --from=site-build ["/app/ui-build/build", "site"]
COPY dist dist

ENTRYPOINT ["/app/backend"]
