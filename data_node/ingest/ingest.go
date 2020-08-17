package ingest

import (
	"encoding/json"
	"fmt"
	"path"

	"github.com/SayedAlesawy/Videra-Storage/config"
	datanode "github.com/SayedAlesawy/Videra-Storage/data_node"
	jobscheduler "github.com/SayedAlesawy/Videra-Storage/data_node/jobs_scheduler"
)

var jobExecutionLoggerPrefix = "[Job-Execution]"

// StartJob starts ingesting file to ingestion module
func StartJob(videoInfo datanode.File) {

	var metadata datanode.VideoMetadata
	json.Unmarshal([]byte(videoInfo.Extras), &metadata)

	var modelInfo datanode.File
	datanode.NodeInstance().DB.Connection.Where("token = ?", metadata.AssociatedModel).Find(&modelInfo)

	var modelExtras datanode.ModelExtras
	json.Unmarshal([]byte(modelInfo.Extras), &modelExtras)

	executeJob(videoInfo.Path, videoInfo.Token, modelInfo.Path, modelExtras.AssociatedConfigPath, modelExtras.AssociatedCodePath, modelInfo.Token, 0, metadata.FramesCount)
}

// executeJob starts command for starting ingestion
func executeJob(videoPath string, videoToken string, modelPath string, configPath string, codePath string, groupID string, startIndex int, framesCount int) {
	command := "./ingestion-engine.bin"
	args := prepareArgs(videoPath, videoToken, modelPath, configPath, codePath, groupID, startIndex, framesCount)
	jobDir := config.ConfigurationManagerInstance("").DataNodeConfig().IngestionModulePath
	jobName := getJobName(groupID)
	jobQueue := jobscheduler.JobQueueInstance()
	jobQueue.InsertJobWithDir(jobName, jobDir, command, args, jobscheduler.PostJob{})
}

func prepareArgs(videoPath string, videoToken string, modelPath string, configPath string, codePath string, groupID string, startIndex int, frameCount int) []string {
	codeFolder, _ := path.Split(codePath)
	execGroupArg := fmt.Sprintf("-execution-group-id=%s", groupID)
	videoPathArg := fmt.Sprintf("-video-path=%s", videoPath)
	videoTokenArg := fmt.Sprintf("-video-token=%s", videoToken)
	modelPathArg := fmt.Sprintf("-model-path=%s", modelPath)
	configPathArg := fmt.Sprintf("-model-config-path=%s", configPath)
	codePathArg := fmt.Sprintf("-code-path=%s", codeFolder)
	startIndexArg := fmt.Sprintf("-start-idx=%d", startIndex)
	frameCountArg := fmt.Sprintf("-frame-count=%d", frameCount)
	args := []string{execGroupArg, videoPathArg, videoTokenArg, modelPathArg, configPathArg, codePathArg, startIndexArg, frameCountArg}
	return args
}

func getJobName(token string) string {
	return fmt.Sprintf("Ingest-%s", token)
}
