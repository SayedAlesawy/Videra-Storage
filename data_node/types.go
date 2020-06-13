package datanode

import "sync"

//DataNode Represents a data storage node in the system
type DataNode struct {
	IP           string       //IP of the data node host
	InternalPort string       //Port on which all internal comm is done
	NameNode     NameNodeData //Houses the needed info about the current name node
}

//NameNodeData Houses the needed info about the name node
type NameNodeData struct {
	IP   string //IP of the name node host
	Port string //Port on which the data node communicates with the name node
}

// UploadManager represents storage to keep files info
// and keeps track of what files are currently in data node
type UploadManager struct {
	fileBase map[string]FileInfo // Holds information about files available in data node
	mux      sync.RWMutex        // For safe concurrent access to filebase
}

// FileInfo represents file information on disk
type FileInfo struct {
	Name        string   // Name of file
	Type        fileType // Indicates type of file (video, model .... etc)
	Path        string   // Path to file (excluding file name), with trailing slash
	Offset      int64    // Offset of bytes to start writing data at
	Size        int64    // Total size of file in bytes
	isCompleted bool     //Indicates if file completed uploading
}

type fileType string

const (
	videoFile  fileType = "VideoFile"
	modelFile           = "ModelFile"
	configFile          = "ConfigFile"
)
