FROM golang:1.24 AS builder

WORKDIR /app

# Cache go mod dependencies
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . ./

ARG APP_VERSION
ARG BUILD_TIME
ARG TARGETOS
ENV GOOS=$TARGETOS
ARG TARGETARCH
ENV GOARCH=$TARGETARCH

# Cache build artifacts
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -ldflags "-extldflags '-static' -X 'github.com/symbioticfi/relay/cmd/utils/root.Version=${APP_VERSION}' -X 'github.com/symbioticfi/relay/cmd/utils/root.BuildTime=${BUILD_TIME}'" -o relay_utils ./cmd/utils && \
    chmod a+x relay_utils

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -ldflags "-extldflags '-static' -X 'github.com/symbioticfi/relay/cmd/relay/root.Version=${APP_VERSION}' -X 'github.com/symbioticfi/relay/cmd/relay/root.BuildTime=${BUILD_TIME}'" -o relay_sidecar ./cmd/relay && \
    chmod a+x relay_sidecar

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/relay_utils .
COPY --from=builder /app/relay_sidecar .
