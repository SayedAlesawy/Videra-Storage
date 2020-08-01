package namenode

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
var logPrefix = "[NameNode]"

// nameNodeOnce Used to garauntee thread safety for singleton instances
var nameNodeOnce sync.Once

// nameNodeInstance A singleton instance of the name node object
var nameNodeInstance *NameNode

// NodeInstance A function to return a singleton name node instance
func NodeInstance() *NameNode {
	configManager := config.ConfigurationManagerInstance("")
	nameNodeConfig := configManager.NameNodeConfig()

	cacheInstance, err := redis.Instance(configManager.RedisConfig())
	errors.HandleError(err, fmt.Sprintf("%s Unable to connect to caching layer", logPrefix), true)

	nameNodeOnce.Do(func() {
		nameNode := NameNode{
			IP:                       nameNodeConfig.IP,
			InternalPort:             nameNodeConfig.InternalRequestsPort,
			dataNodesTrackingKey:     nameNodeConfig.DataNodesTrackingKey,
			dataNodeOfflineThreshold: nameNodeConfig.DataNodeOfflineThreshold,
			InteralReqTimeout:        time.Duration(nameNodeConfig.InternalReqTimeout) * time.Second,
			HealthCheckInterval:      time.Duration(nameNodeConfig.HealthCheckInterval) * time.Second,
			cache:                    cacheInstance,
			DB:                       database.DBInstance(nameNodeConfig.StorageDBName),
		}

		nameNode.DB.Connection.AutoMigrate(&Clip{})

		nameNodeInstance = &nameNode
	})

	return nameNodeInstance
}
