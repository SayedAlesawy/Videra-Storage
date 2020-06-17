package main

import (
	datanode "github.com/SayedAlesawy/Videra-Storage/data_node"
	"github.com/SayedAlesawy/Videra-Storage/data_node/controllers/inner"
	"github.com/SayedAlesawy/Videra-Storage/data_node/controllers/outer"
)

func main() {
	dataNode := datanode.NodeInstance()
	dataNode.JoinCluster()

	go inner.ServerInstance().Start()

	outer.ServerInstance().Start()
}
