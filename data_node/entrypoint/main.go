package main

import (
	datanode "github.com/SayedAlesawy/Videra-Storage/data_node"
	"github.com/SayedAlesawy/Videra-Storage/data_node/controllers/inner"
)

func main() {
	dataNode := datanode.NodeInstance()
	dataNode.JoinCluster()

	go inner.ServerInstance().Start()

	um := datanode.UploadManagerInstance()
	um.Start()
}
