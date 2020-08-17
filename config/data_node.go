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
	GPU                          string //Indicates if the datanode has a GPU
	NetworkProtocol              string //Network protocol used by the data node
	NameNodeReplicationURL       string //URL to request datanodes for replication
	StorageDBName                string //Storage database name
	InternalReqTimeout           int    //Timeout for internal requests
	MaxRequestSize               int64  //Maximum acceptable size of body size
	RejoinClusterInterval        int    //Freqency of retrying the join cluster request
	MetadataCommand              string //Command for running script to fetch video metadata
	MetadataScriptPath           string //Path to fetch metadata script
	IngestionModulePath          string //Path to ingestion module to execute jobs
	ReplicationNumberOfRetries   int    //Number of retries when a failure happens in replication
	ReplicationWaitingTime       int    //Waiting time between failed retries in replication
	StreamOutputVideoWidth       int    //Width of streaming output video
	StreamOutputVideoHeight      int    //Height of streaming output video
	StreamSegmentTime            int    //Segment time in seconds for HLS protocol
	StreamPlaylistName           string //HLS playlist file name
	StreamFolderName             string //Name of folder that contains streaming files
	ThumbnailOutputWidth         int    //Width of streaming output video
	ThumbnailOutputHeight        int    //Height of streaming output video
	ThumbnailCaptureSecond       int    //Time to capture thumbnail at, in seconds
	ThumbnailFolderName          string //Name of folder that contains thumbnail files
	MaximumConcurrentJobs        int    //Maximum number of running concurrent jobs
	JobTimeout                   int    //Maximum time for a job untill timeout, in seconds
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
			GPU:                          envString("GPU_STATUS", "false"),
			NameNodeReplicationURL:       envString("NAME_NODE_REPLICATION_URL", "http://localhost:8080/replication"),
			NetworkProtocol:              envString("NET_PROTOCOL", "tcp"),
			StorageDBName:                envString("STO_DB_NAME", "videra_storage"),
			InternalReqTimeout:           int(envInt("INTERNAL_REQ_TIMEOUT", "5")),
			MaxRequestSize:               envInt("MAX_REQUEST_SIZE", "4194304"),
			RejoinClusterInterval:        int(envInt("REJOIN_CLUSTER_INTERVAL", "2")),
			MetadataCommand:              envString("METADATA_COMMAND", "/usr/bin/python3"),
			MetadataScriptPath:           envString("METADATA_SCRIPT", "../../scripts/fetch_metadata.py"),
			IngestionModulePath:          envString("INGESTION_MODULE_PATH", "/home/ahmed/Downloads/Videra-Ingestion/orchestrator"),
			ReplicationNumberOfRetries:   int(envInt("REPLICATION_RETIRES", "3")),
			ReplicationWaitingTime:       int(envInt("REPLICATION_WAITING_TIME", "5")),
			StreamOutputVideoWidth:       int(envInt("STREAM_VIDEO_WIDTH", "256")),
			StreamOutputVideoHeight:      int(envInt("STREAM_VIDEO_HEIGHT", "144")),
			StreamSegmentTime:            int(envInt("STREAM_SEGMENT_TIME", "60")),
			StreamPlaylistName:           envString("STREAM_PLAYLIST_NAME", "index.m3u8"),
			StreamFolderName:             envString("STREAM_FOLDER_NAME", "stream"),
			ThumbnailCaptureSecond:       int(envInt("THUMBNAIL_CAPTURE_SECOND", "5")),
			ThumbnailOutputWidth:         int(envInt("THUMBNAIL_OUTPUT_WIDTH", "256")),
			ThumbnailOutputHeight:        int(envInt("THUMBNAIL_OUTPUT_HEIGHT", "144")),
			ThumbnailFolderName:          envString("THUMBNAIL_FOLDER_NAME", "thumbnail"),
			MaximumConcurrentJobs:        int(envInt("MAXIMUM_CONCURRENT_JOBS", "1")),
			JobTimeout:                   int(envInt("JOB_TIMEOUT", "7200")),
		}

		dataNodeConfigInstance = &dataNodeConfig
	})

	return dataNodeConfigInstance
}
