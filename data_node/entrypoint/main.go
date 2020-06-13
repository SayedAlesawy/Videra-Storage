package main

import (
	datanode "github.com/SayedAlesawy/Videra-Storage/data_node"
)

func main() {
	dataNode := datanode.NodeInstance()
	dataNode.JoinCluster()
}
