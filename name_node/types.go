package namenode

import (
	"time"
)

// NameNode Represents a tracking node in the storage system
type NameNode struct {
	IP                  string         //IP of the name node host
	InternalPort        string         //Port on which all internal comm is done
	InteralReqTimeout   time.Duration  //Timeout for internal requests
	HealthCheckInterval time.Duration  //The frequency of the health check request to data nodes
	DataNodes           []DataNodeData //Array of all tracked data nodes
}

//DataNodeData Houses the needed info about a data node
type DataNodeData struct {
	IP   string //IP of the data node host
	Port string //Port on which the name node communicates with the data node
}
