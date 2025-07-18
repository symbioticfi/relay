FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 go build -ldflags "-extldflags '-static'" -o utils ./cmd/relay_utils && \
    		chmod a+x relay_utils
RUN CGO_ENABLED=0 go build -ldflags "-extldflags '-static'" -o relay ./cmd/relay_sidecar && \
    		chmod a+x relay_sidecar

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/relay_utils .
COPY --from=builder /app/relay_sidecar .
COPY --from=builder /app/circuits /app/circuits
