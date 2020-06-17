package datanode

import (
	"time"
)

//DataNode Represents a data storage node in the system
type DataNode struct {
	IP                    string        //IP of the data node host
	ID                    string        //Unique ID for the data node
	Port                  string        //Port on which all external comm is done
	InternalPort          string        //Port on which all internal comm is done
	InternalReqTimeout    time.Duration //Timeout for internal requests
	RejoinClusterInterval time.Duration //Frequency of the rejoin cluster request
	NameNode              NameNodeData  //Houses the needed info about the current name node
}

//NameNodeData Houses the needed info about the name node
type NameNodeData struct {
	IP   string //IP of the name node host
	Port string //Port on which the data node communicates with the name node
}

// FileInfo represents file information on disk
type FileInfo struct {
	Name        string   // Name of file
	Type        fileType // Indicates type of file (video, model .... etc)
	Path        string   // Path to file (excluding file name)
	Offset      int64    // Offset of bytes to start writing data at
	Size        int64    // Total size of file in bytes
	IsCompleted bool     //Indicates if file completed uploading
}

type fileType string

const (
	videoFile  fileType = "VideoFile"
	modelFile           = "ModelFile"
	configFile          = "ConfigFile"
)
