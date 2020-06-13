package config

import (
	"sync"
)

// NameNodeconfig Houses the configurations of the name node
type NameNodeconfig struct {
	IP                   string //Name node IP
	InternalRequestsPort string //The internal requests ports
	NetowrkProtocol      string //Name network protcol
}

// nameNodeConfigOnce Used to garauntee thread safety for singleton instances
var nameNodeConfigOnce sync.Once

// nameNodeConfigInstance A singleton instance of the name node object
var nameNodeConfigInstance *NameNodeconfig

// NameNodeConfig A function to name node configs
func (manager *ConfigurationManager) NameNodeConfig() *NameNodeconfig {
	nameNodeConfigOnce.Do(func() {
		nameNodeConfig := NameNodeconfig{
			IP:                   envString("IP", "127.0.0.1"),
			InternalRequestsPort: envString("INTERNAL_REQ_PORT", "7000"),
			NetowrkProtocol:      envString("NET_PROTOCOL", "tcp"),
		}

		nameNodeConfigInstance = &nameNodeConfig
	})

	return nameNodeConfigInstance
}
