package namenode

import (
	"time"

	"github.com/SayedAlesawy/Videra-Storage/drivers/redis"
	"github.com/SayedAlesawy/Videra-Storage/utils/database"
)

// NameNode Represents a tracking node in the storage system
type NameNode struct {
	IP                       string             //IP of the name node host
	InternalPort             string             //Port on which all internal comm is done
	dataNodesTrackingKey     string             //The key of the redis hash used to track data nodes
	dataNodeOfflineThreshold int                //Threshold of missed pings at which a data node is considered offline
	InteralReqTimeout        time.Duration      //Timeout for internal requests
	HealthCheckInterval      time.Duration      //The frequency of the health check request to data nodes
	DataNodes                []DataNodeData     //Array of all tracked data nodes
	cache                    *redis.Client      //Used by the name node to access a persistent caching layer
	DB                       *database.Database //Database connection
}

//DataNodeData Houses the needed info about a data node
type DataNodeData struct {
	ID           string `json:"id"`            //Unique ID for each data node
	IP           string `json:"ip"`            //IP of the data node host
	InternalPort string `json:"internal_port"` //Port on which the name node communicates with the data node
	Port         string `json:"port"`          //Port on which the data node communicates with clients
	Latency      int    `json:"latency"`       //Count of missed pings by the data node
	RequestCount uint   `json:"request_count"` //Number of clients requests routed to the data node
}
