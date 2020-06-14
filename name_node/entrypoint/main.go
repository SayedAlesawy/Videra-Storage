package main

import (
	namenode "github.com/SayedAlesawy/Videra-Storage/name_node"
	"github.com/SayedAlesawy/Videra-Storage/name_node/controllers/inner"
)

func main() {
	nameNode := namenode.NodeInstance()

	go nameNode.PingDataNodes()

	inner.ServerInstance().Start()
}
