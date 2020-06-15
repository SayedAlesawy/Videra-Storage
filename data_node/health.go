package datanode

import (
	"context"
	"fmt"
	"log"

	"github.com/SayedAlesawy/Videra-Storage/name_node/nnpb"
	"github.com/SayedAlesawy/Videra-Storage/utils/errors"
	"google.golang.org/grpc"
)

// JoinCluster A function to notify the name node to join the cluster
func (dataNode *DataNode) JoinCluster() {
	conn, err := grpc.Dial(dataNode.getNameNodeAddress(), grpc.WithBlock(), grpc.WithInsecure())
	errors.HandleError(err, fmt.Sprintf("%s Unable to connect to name node", logPrefix), true)
	defer conn.Close()

	client := nnpb.NewNameNodeInternalRoutesClient(conn)

	log.Println(logPrefix, "Sending join cluster request to name node")
	req := nnpb.JoinClusterRequest{
		IP:   dataNode.IP,
		Port: dataNode.InternalPort,
	}

	ctx, cancel := context.WithTimeout(context.Background(), dataNode.InteralReqTimeout)
	defer cancel()

	joinStatus, err := client.JoinCluster(ctx, &req)
	errors.HandleError(err, fmt.Sprintf("%s %v.JoinCluster(_) = _, %v: ", logPrefix, client, err), false)

	log.Println(logPrefix, "Server responded with: ", joinStatus.Status)
}
