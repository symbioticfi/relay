package v1

import (
	"github.com/symbioticfi/relay/internal/gen/api/v1"
	"google.golang.org/grpc"
)

// Client wraps the generated gRPC client
type Client struct {
	v1.SymbioticAPIServiceClient
}

// NewClient creates a new client instance for symbiotic relay
func NewClient(conn grpc.ClientConnInterface) *Client {
	return &Client{
		v1.NewSymbioticAPIServiceClient(conn),
	}
}
