package datanode

import (
	"sync"
	"time"

	"github.com/SayedAlesawy/Videra-Storage/config"
	"github.com/SayedAlesawy/Videra-Storage/utils/database"
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
			IP:                    dataNodeConfig.IP,
			ID:                    dataNodeConfig.ID,
			Port:                  dataNodeConfig.Port,
			InternalPort:          dataNodeConfig.InternalRequestsPort,
			InternalReqTimeout:    time.Duration(dataNodeConfig.InternalReqTimeout) * time.Second,
			RejoinClusterInterval: time.Duration(dataNodeConfig.RejoinClusterInterval) * time.Second,
			NameNode: NameNodeData{
				IP:   dataNodeConfig.NameNodeIP,
				Port: dataNodeConfig.NameNodeInternalRequestsPort,
			},
			DB: database.DBInstance(dataNodeConfig.StorageDBName),
		}

		dataNode.DB.Connection.AutoMigrate(&File{})

		dataNodeInstance = &dataNode
	})

	return dataNodeInstance
}
