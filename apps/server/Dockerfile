# syntax=docker/dockerfile:1.7@sha256:a57df69d0ea827fb7266491f2813635de6f17269be881f696fbfdf2d83dda33e
FROM --platform=$BUILDPLATFORM golang:1.25.14-alpine@sha256:1ae0735f00daffa3aaf1363a5184c0d2dc55c78e3db4ec70241cdac97bf84b59 AS builder

ARG TARGETOS
ARG TARGETARCH
ARG BUILD_VERSION=development

WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    --mount=type=cache,target=/root/.cache/go-build,sharing=locked \
    go mod download

COPY . .

RUN --mount=type=cache,target=/root/.cache/go-build,sharing=locked \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags "-X main.version=${BUILD_VERSION}" -o /app/bin/api ./cmd/api \
    && CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags "-X main.version=${BUILD_VERSION}" -o /app/bin/worker ./cmd/worker

FROM alpine:3.24@sha256:28bd5fe8b56d1bd048e5babf5b10710ebe0bae67db86916198a6eec434943f8b AS runtime

ARG BUILD_VERSION=development

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S -g 10001 fortyone \
    && adduser -S -D -H -u 10001 -G fortyone fortyone

WORKDIR /app

COPY --from=builder --chown=fortyone:fortyone /app/templates/ /app/templates/

USER fortyone

FROM runtime AS server

LABEL org.opencontainers.image.title="FortyOne API" \
    org.opencontainers.image.revision="$BUILD_VERSION"

COPY --from=builder --chown=fortyone:fortyone /app/bin/api /app/api

EXPOSE 8000

CMD ["/app/api"]

FROM runtime AS worker

LABEL org.opencontainers.image.title="FortyOne worker" \
    org.opencontainers.image.revision="$BUILD_VERSION"

COPY --from=builder --chown=fortyone:fortyone /app/bin/worker /app/worker

EXPOSE 8080

CMD ["/app/worker"]
