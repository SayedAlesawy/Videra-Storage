package datanode

import (
	"fmt"
	"sync"
)

// logPrefix Used for hierarchical logging
var logPrefix = "[DataNode]"

// dataNodeOnce Used to garauntee thread safety for singleton instances
var dataNodeOnce sync.Once

// dataNodeInstance A singleton instance of the data node object
var dataNodeInstance *DataNode

// NodeInstance A function to return a singleton data node instance
func NodeInstance() *DataNode {
	dataNodeOnce.Do(func() {
		dataNode := DataNode{
			IP:           "127.0.0.1",
			InternalPort: "4444",
			NameNode: NameNodeData{
				IP:   "127.0.0.1",
				Port: "5555",
			},
		}

		dataNodeInstance = &dataNode
	})

	return dataNodeInstance
}

// getNameNodeAddress A function to get the name node address
func (dataNode *DataNode) getNameNodeAddress() string {
	return fmt.Sprintf("%s:%s", dataNode.NameNode.IP, dataNode.NameNode.Port)
}
