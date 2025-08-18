FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

ARG APP_VERSION
ARG BUILD_TIME
ARG TARGETOS
ENV GOOS=$TARGETOS
ARG TARGETARCH
ENV GOARCH=$TARGETARCH

RUN CGO_ENABLED=0 go build -ldflags "-extldflags '-static' -X 'main.Version=${APP_VERSION}' -X 'main.BuildTime=${BUILD_TIME}'" -o relay_utils ./cmd/utils && \
    		chmod a+x relay_utils
RUN CGO_ENABLED=0 go build -ldflags "-extldflags '-static' -X 'main.Version=${APP_VERSION}' -X 'main.BuildTime=${BUILD_TIME}'" -o relay_sidecar ./cmd/relay && \
    		chmod a+x relay_sidecar

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/relay_utils .
COPY --from=builder /app/relay_sidecar .
