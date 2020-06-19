package datanode

import (
	"time"

	"github.com/SayedAlesawy/Videra-Storage/utils/database"
)

//DataNode Represents a data storage node in the system
type DataNode struct {
	IP                    string             //IP of the data node host
	ID                    string             //Unique ID for the data node
	Port                  string             //Port on which all external comm is done
	InternalPort          string             //Port on which all internal comm is done
	InternalReqTimeout    time.Duration      //Timeout for internal requests
	RejoinClusterInterval time.Duration      //Frequency of the rejoin cluster request
	NameNode              NameNodeData       //Houses the needed info about the current name node
	DB                    *database.Database //Database connection
}

//NameNodeData Houses the needed info about the name node
type NameNodeData struct {
	IP   string //IP of the name node host
	Port string //Port on which the data node communicates with the name node
}

const (
	//VideoFileType represents video type
	VideoFileType string = "video"
	//ModelFileType represents model type
	ModelFileType string = "model"
	//ConfigFileType represents config type
	ConfigFileType string = "config"
)

//SupportedFileTypes represents list of supported file types
var SupportedFileTypes = [...]string{VideoFileType, ModelFileType, ConfigFileType}
