package datanode

import (
	"time"

	"github.com/SayedAlesawy/Videra-Storage/drivers/redis"
	"github.com/SayedAlesawy/Videra-Storage/utils/database"
)

//DataNode Represents a data storage node in the system
type DataNode struct {
	IP                    string             //IP of the data node host
	ID                    string             //Unique ID for the data node
	Port                  string             //Port on which all external comm is done
	InternalPort          string             //Port on which all internal comm is done
	GPU                   bool               //Indicates if the data node has a GPU
	InternalReqTimeout    time.Duration      //Timeout for internal requests
	RejoinClusterInterval time.Duration      //Frequency of the rejoin cluster request
	NameNode              NameNodeData       //Houses the needed info about the current name node
	DB                    *database.Database //Database connection
	Cache                 *redis.Client      //Used by the data node to access a persistent caching layer
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
)

//SupportedFileTypes represents list of supported file types
var SupportedFileTypes = [...]string{VideoFileType, ModelFileType}

// ModelExtras represents extra parameters associated with model
type ModelExtras struct {
	ModelSize            int64  `json:"model_size"`             //size of model file
	AssociatedConfigPath string `json:"associated_config_path"` //path to config file
	AssociatedConfigSize int64  `json:"associated_config_size"` //config file size
	AssociatedCodePath   string `json:"associated_code_path"`   //path to code file
	AssociatedCodeSize   int64  `json:"associated_code_size"`   //code file size
}

//VideoMetadata Represents video metadata model
type VideoMetadata struct {
	Height          int     `json:"height"`           //Height of video
	Width           int     `json:"width"`            //Width of video
	FramesCount     int     `json:"frames_count"`     //Number of frames in video
	Fps             float64 `json:"fps"`              //Frames per second
	Duration        float64 `json:"duration"`         //Length of video
	AssociatedModel string  `json:"associated_model"` //Associated Model ID
}
