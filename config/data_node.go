package config

import (
	"sync"
)

// DataNodeconfig Houses the configurations of the data node
type DataNodeconfig struct {
	IP                           string //Name node IP
	NameNodeIP                   string //IP of the current name node
	InternalRequestsPort         string //The internal requests ports
	NameNodeInternalRequestsPort string //The internal requests port of the name node
	Port                         string //Port to listen to requests
	InternalReqTimeout           int    //Timeout for internal requests
}

// dataNodeConfigOnce Used to garauntee thread safety for singleton instances
var dataNodeConfigOnce sync.Once

// dataNodeConfigInstance A singleton instance of the data node object
var dataNodeConfigInstance *DataNodeconfig

// DataNodeConfig A function to data node configs
func (manager *ConfigurationManager) DataNodeConfig() *DataNodeconfig {
	dataNodeConfigOnce.Do(func() {
		dataNodeConfig := DataNodeconfig{
			IP:                           envString("IP", "127.0.0.1"),
			NameNodeIP:                   envString("NAME_NODE_IP", "127.0.0.1"),
			InternalRequestsPort:         envString("INTERNAL_REQ_PORT", "6000"),
			NameNodeInternalRequestsPort: envString("NAME_NODE_INTERNAL_REQ_PORT", "7000"),
			Port:                         envString("PORT", "8080"),
			InternalReqTimeout:           int(envInt("INTERNAL_REQ_TIMEOUT", "5")),
		}

		dataNodeConfigInstance = &dataNodeConfig
	})

	return dataNodeConfigInstance
}
