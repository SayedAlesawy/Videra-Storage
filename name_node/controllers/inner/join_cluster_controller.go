package inner

import (
	context "context"
	"log"

	"github.com/SayedAlesawy/Videra-Storage/ndpb"
)

// JoinCluster Handles the join cluster request
func (server *Server) JoinCluster(ctx context.Context, req *ndpb.JoinClusterRequest) (*ndpb.JoinClusterResponse, error) {
	log.Println(logPrefix, "Received:", "IP:", req.IP, "Port:", req.Port)

	return &ndpb.JoinClusterResponse{
		Status: ndpb.JoinClusterResponse_SUCCESS,
	}, nil
}
