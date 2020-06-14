package inner

import (
	context "context"
	"log"

	"github.com/SayedAlesawy/Videra-Storage/name_node/nnpb"
)

// JoinCluster Handles the join cluster request
func (server *Server) JoinCluster(ctx context.Context, req *nnpb.JoinClusterRequest) (*nnpb.JoinClusterResponse, error) {
	log.Println(logPrefix, "Received:", "IP:", req.IP, "Port:", req.Port)

	return &nnpb.JoinClusterResponse{
		Status: nnpb.JoinClusterResponse_SUCCESS,
	}, nil
}
