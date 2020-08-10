package datanode

import (
	"fmt"
	"sync"
	"time"

	"github.com/SayedAlesawy/Videra-Storage/config"
	"github.com/SayedAlesawy/Videra-Storage/drivers/redis"
	"github.com/SayedAlesawy/Videra-Storage/utils/database"
	"github.com/SayedAlesawy/Videra-Storage/utils/errors"
)

// logPrefix Used for hierarchical logging
var logPrefix = "[DataNode]"

// dataNodeOnce Used to garauntee thread safety for singleton instances
var dataNodeOnce sync.Once

// dataNodeInstance A singleton instance of the data node object
var dataNodeInstance *DataNode

// NodeInstance A function to return a singleton data node instance
func NodeInstance() *DataNode {
	configManager := config.ConfigurationManagerInstance("")
	dataNodeConfig := configManager.DataNodeConfig()
	cacheInstance, err := redis.Instance(configManager.RedisConfig())
	errors.HandleError(err, fmt.Sprintf("%s Unable to connect to caching layer", logPrefix), true)

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
			DB:    database.DBInstance(dataNodeConfig.StorageDBName),
			Cache: cacheInstance,
		}

		dataNode.DB.Connection.AutoMigrate(&File{})

		dataNodeInstance = &dataNode
	})

	return dataNodeInstance
}
