package v1

import (
	apiv1 "github.com/symbioticfi/relay/internal/gen/api/v1"
	"google.golang.org/grpc"
)

// SymbioticClient wraps the generated gRPC client
type SymbioticClient struct {
	apiv1.SymbioticAPIServiceClient
}

// NewSymbioticClient creates a new client instance for symbiotic relay
func NewSymbioticClient(conn grpc.ClientConnInterface) *SymbioticClient {
	return &SymbioticClient{
		apiv1.NewSymbioticAPIServiceClient(conn),
	}
}
