package inner

import (
	context "context"
	"log"

	"github.com/SayedAlesawy/Videra-Storage/data_node/dnpb"
)

// HealthCheck Handles the join cluster request
func (server *Server) HealthCheck(ctx context.Context, req *dnpb.HealthCheckRequest) (*dnpb.HealthCheckResponse, error) {
	log.Println(logPrefix, "Received health check ping from name node")

	return &dnpb.HealthCheckResponse{
		Status: dnpb.HealthCheckResponse_HEALTHY,
	}, nil
}
