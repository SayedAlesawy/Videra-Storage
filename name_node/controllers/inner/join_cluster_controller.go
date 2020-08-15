package inner

import (
	context "context"
	"fmt"
	"log"

	namenode "github.com/SayedAlesawy/Videra-Storage/name_node"
	"github.com/SayedAlesawy/Videra-Storage/name_node/nnpb"
)

// JoinCluster Handles the join cluster request
func (server *Server) JoinCluster(ctx context.Context, req *nnpb.JoinClusterRequest) (*nnpb.JoinClusterResponse, error) {
	log.Println(logPrefix, fmt.Sprintf("Received join cluster from node: %s on %s:%s", req.ID, req.IP, req.InternalPort))

	dataNodeData := namenode.NewDataNodeData(req.ID, req.IP, req.InternalPort, req.Port, req.GPU)

	ok := namenode.NodeInstance().InsertDataNodeData(dataNodeData)
	var status nnpb.JoinClusterResponse_JoinStatus
	if ok {
		status = nnpb.JoinClusterResponse_SUCCESS
	} else {
		status = nnpb.JoinClusterResponse_FAILURE
	}

	return &nnpb.JoinClusterResponse{
		Status: status,
	}, nil
}
