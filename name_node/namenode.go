package namenode

import (
	"sync"
	"time"

	"github.com/SayedAlesawy/Videra-Storage/config"
)

// logPrefix Used for hierarchical logging
var logPrefix = "[NameNode]"

// nameNodeOnce Used to garauntee thread safety for singleton instances
var nameNodeOnce sync.Once

// nameNodeInstance A singleton instance of the name node object
var nameNodeInstance *NameNode

// NodeInstance A function to return a singleton name node instance
func NodeInstance() *NameNode {
	nameNodeConfig := config.ConfigurationManagerInstance("").NameNodeConfig()

	nameNodeOnce.Do(func() {
		dataNode := NameNode{
			IP:                  nameNodeConfig.IP,
			InternalPort:        nameNodeConfig.InternalRequestsPort,
			InteralReqTimeout:   time.Duration(nameNodeConfig.InternalReqTimeout) * time.Second,
			HealthCheckInterval: time.Duration(nameNodeConfig.HealthCheckInterval) * time.Second,
			DataNodes: []DataNodeData{
				{IP: "127.0.0.1", Port: "6001"},
				{IP: "127.0.0.1", Port: "6002"},
				{IP: "127.0.0.1", Port: "6003"},
			},
		}

		nameNodeInstance = &dataNode
	})

	return nameNodeInstance
}
