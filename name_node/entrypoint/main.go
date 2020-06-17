package main

import (
	namenode "github.com/SayedAlesawy/Videra-Storage/name_node"
	"github.com/SayedAlesawy/Videra-Storage/name_node/controllers/inner"
	"github.com/SayedAlesawy/Videra-Storage/name_node/controllers/outer"
)

func main() {
	nameNode := namenode.NodeInstance()

	go nameNode.PingDataNodes()

	go inner.ServerInstance().Start()

	outer.ServerInstance().Start()
}
