FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 go build -ldflags "-extldflags '-static'" -o generate_genesis ./cmd/generate-genesis && \
    		chmod a+x generate_genesis
RUN CGO_ENABLED=0 go build -ldflags "-extldflags '-static'" -o middleware_offchain ./cmd/middleware-offchain && \
    		chmod a+x middleware_offchain
RUN CGO_ENABLED=0 go build -ldflags "-extldflags '-static'" -o msg_sign ./cmd/msg-sign && \
    		chmod a+x msg_sign

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/middleware_offchain .
COPY --from=builder /app/msg_sign .
COPY --from=builder /app/circuits /app/circuits
