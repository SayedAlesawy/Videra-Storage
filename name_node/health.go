package namenode

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/SayedAlesawy/Videra-Ingestion/orchestrator/utils/errors"
	"github.com/SayedAlesawy/Videra-Storage/data_node/dnpb"
	"google.golang.org/grpc"
)

// PingDataNodes A function to ping all currently conneced data nodes for health checking
func (nameNode *NameNode) PingDataNodes() {
	time.Sleep(10 * time.Second)

	for range time.Tick(nameNode.HealthCheckInterval) {
		for _, dataNode := range nameNode.DataNodes {
			address := nameNode.getNameNodeAddress(dataNode)

			conn, err := grpc.Dial(address, grpc.WithInsecure())
			defer conn.Close()
			if errors.IsError(err) {
				log.Println(fmt.Sprintf("%s Unable to connect to data node on: %s", logPrefix, address))
				continue
			}

			client := dnpb.NewDataNodeInternalRoutesClient(conn)
			req := dnpb.HealthCheckRequest{}

			ctx, cancel := context.WithTimeout(context.Background(), nameNode.InteralReqTimeout)
			defer cancel()

			healthCheckResp, err := client.HealthCheck(ctx, &req)
			if errors.IsError(err) {
				//TODO: Remove it from list of tracked nodes
				log.Println(fmt.Sprintf("%s Data node on address: %s is OFFLINE", logPrefix, address))
				continue
			}

			log.Println(logPrefix, fmt.Sprintf("Data node on address: %s is:", address), healthCheckResp.Status)
		}
	}
}
