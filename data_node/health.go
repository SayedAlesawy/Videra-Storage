package datanode

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/SayedAlesawy/Videra-Ingestion/orchestrator/utils/errors"
	"github.com/SayedAlesawy/Videra-Storage/ndpb"
	"google.golang.org/grpc"
)

// JoinCluster A function to notify the name node to join the cluster
func (dataNode *DataNode) JoinCluster() {
	conn, err := grpc.Dial(dataNode.getNameNodeAddress(), grpc.WithBlock(), grpc.WithInsecure())
	errors.HandleError(err, fmt.Sprintf("%s Unable to connect to name node", logPrefix), true)
	defer conn.Close()

	client := ndpb.NewNameNodeInternalRoutesClient(conn)

	log.Println(logPrefix, "Sending join cluster request to name node")
	req := ndpb.JoinClusterRequest{
		IP:   dataNode.IP,
		Port: dataNode.InternalPort,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	joinStatus, err := client.JoinCluster(ctx, &req)
	errors.HandleError(err, fmt.Sprintf("%s %v.JoinCluster(_) = _, %v: ", logPrefix, client, err), false)

	log.Println(logPrefix, "Server responded with: ", joinStatus.Status)
}
