package config

import (
	"sync"
)

// DataNodeconfig Houses the configurations of the data node
type DataNodeconfig struct {
	IP                           string //Name node IP
	ID                           string //Unique ID for the data node
	NameNodeIP                   string //IP of the current name node
	InternalRequestsPort         string //The internal requests ports
	NameNodeInternalRequestsPort string //The internal requests port of the name node
	Port                         string //Port to listen to requests
	NetworkProtocol              string //Network protocol used by the data node
	StorageDBName                string //Storage database name
	InternalReqTimeout           int    //Timeout for internal requests
	MaxRequestSize               int64  //Maximum acceptable size of body size
	RejoinClusterInterval        int    //Freqency of retrying the join cluster request
	MetadataCommand              string //Command for running script to fetch video metadata
	MetadataScriptPath           string //Path to fetch metadata script
	IngestionModuleCommand       string //Command to run Ingestion Module
	IngestionModulePath          string //Path to ingestion module to execute jobs
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
			ID:                           envString("ID", "1"),
			NameNodeIP:                   envString("NAME_NODE_IP", "127.0.0.1"),
			InternalRequestsPort:         envString("INTERNAL_REQ_PORT", "6000"),
			NameNodeInternalRequestsPort: envString("NAME_NODE_INTERNAL_REQ_PORT", "7000"),
			Port:                         envString("PORT", "8080"),
			NetworkProtocol:              envString("NET_PROTOCOL", "tcp"),
			StorageDBName:                envString("STO_DB_NAME", "videra_storage"),
			InternalReqTimeout:           int(envInt("INTERNAL_REQ_TIMEOUT", "5")),
			MaxRequestSize:               envInt("MAX_REQUEST_SIZE", "4194304"),
			RejoinClusterInterval:        int(envInt("REJOIN_CLUSTER_INTERVAL", "2")),
			MetadataCommand:              envString("METADATA_COMMAND", "/usr/bin/python3"),
			MetadataScriptPath:           envString("METADATA_SCRIPT", "../../scripts/fetch_metadata.py"),
			IngestionModuleCommand:       envString("INGESTION_MODULE_COMMAND", "make"),
			IngestionModulePath:          envString("INGESTION_MODULE_PATH", "/home/ahmed/Downloads/Videra-Ingestion/orchestrator"),
		}

		dataNodeConfigInstance = &dataNodeConfig
	})

	return dataNodeConfigInstance
}
