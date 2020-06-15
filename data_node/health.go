package datanode

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/SayedAlesawy/Videra-Storage/name_node/nnpb"
	"github.com/SayedAlesawy/Videra-Storage/utils/errors"
	"google.golang.org/grpc"
)

// JoinCluster A function to notify the name node to join the cluster
func (dataNode *DataNode) JoinCluster() {
	for range time.Tick(dataNode.RejoinClusterInterval) {
		conn, err := grpc.Dial(dataNode.getNameNodeAddress(), grpc.WithBlock(), grpc.WithInsecure())
		errors.HandleError(err, fmt.Sprintf("%s Unable to connect to name node", logPrefix), true)
		defer conn.Close()

		client := nnpb.NewNameNodeInternalRoutesClient(conn)

		log.Println(logPrefix, "Sending join cluster request to name node")
		req := nnpb.JoinClusterRequest{
			ID:   dataNode.ID,
			IP:   dataNode.IP,
			Port: dataNode.InternalPort,
		}

		ctx, cancel := context.WithTimeout(context.Background(), dataNode.InteralReqTimeout)
		defer cancel()

		joinStatus, err := client.JoinCluster(ctx, &req)
		if errors.IsError(err) {
			log.Println(logPrefix, "Unable to join cluster")
			conn.Close()

			continue
		}

		if joinStatus.Status == nnpb.JoinClusterResponse_SUCCESS {
			log.Println(logPrefix, "Successfully joined cluster")

			return
		}

		log.Println(logPrefix, "Unable to join cluster")
		conn.Close()
	}
}
