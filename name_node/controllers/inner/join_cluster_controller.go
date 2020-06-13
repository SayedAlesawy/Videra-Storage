package inner

import (
	context "context"
	"log"

	"github.com/SayedAlesawy/Videra-Storage/protobuf"
)

// JoinCluster Handles the join cluster request
func (server *Server) JoinCluster(ctx context.Context, req *protobuf.JoinClusterRequest) (*protobuf.JoinClusterResponse, error) {
	log.Println("Received:", "IP:", req.IP, "Port:", req.Port)

	return &protobuf.JoinClusterResponse{
		Status: protobuf.JoinClusterResponse_SUCCESS,
	}, nil
}
