package datanode

import (
	"fmt"
	"sync"

	"github.com/SayedAlesawy/Videra-Storage/config"
)

// logPrefix Used for hierarchical logging
var logPrefix = "[DataNode]"

// dataNodeOnce Used to garauntee thread safety for singleton instances
var dataNodeOnce sync.Once

// dataNodeInstance A singleton instance of the data node object
var dataNodeInstance *DataNode

// NodeInstance A function to return a singleton data node instance
func NodeInstance() *DataNode {
	dataNodeConfig := config.ConfigurationManagerInstance("").DataNodeConfig()

	dataNodeOnce.Do(func() {
		dataNode := DataNode{
			IP:           dataNodeConfig.IP,
			InternalPort: dataNodeConfig.InternalRequestsPort,
			NameNode: NameNodeData{
				IP:   dataNodeConfig.NameNodeIP,
				Port: dataNodeConfig.NameNodeInternalRequestsPort,
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
