package v1

import (
	"context"

	votingpowerv1 "github.com/symbioticfi/relay/internal/gen/votingpower/v1"
	"google.golang.org/grpc"
)

type Server struct {
	votingpowerv1.UnimplementedVotingPowerProviderServiceServer
}

func NewServer() *Server {
	return &Server{}
}

func RegisterVotingPowerProviderServiceServer(registrar grpc.ServiceRegistrar, srv votingpowerv1.VotingPowerProviderServiceServer) {
	votingpowerv1.RegisterVotingPowerProviderServiceServer(registrar, srv)
}

func (s *Server) GetVotingPowersAt(_ context.Context, _ *votingpowerv1.GetVotingPowersAtRequest) (*votingpowerv1.GetVotingPowersAtResponse, error) {
	return &votingpowerv1.GetVotingPowersAtResponse{
		VotingPowers: []*votingpowerv1.OperatorVotingPower{},
	}, nil
}
