package outer

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/SayedAlesawy/Videra-Storage/config"
	datanode "github.com/SayedAlesawy/Videra-Storage/data_node"
	"github.com/SayedAlesawy/Videra-Storage/utils/errors"
)

var jobExecutionLoggerPrefix = "[Job-Execution]"

func startJob(videoInfo datanode.File) {

	var metadata datanode.VideoMetadata
	json.Unmarshal([]byte(videoInfo.Extras), &metadata)

	var modelInfo datanode.File
	datanode.NodeInstance().DB.Connection.Where("token = ?", metadata.AssociatedModel).Find(&modelInfo)

	var modelExtras datanode.ModelExtras
	json.Unmarshal([]byte(modelInfo.Extras), &modelExtras)

	executeJob(videoInfo.Path, modelInfo.Path, modelExtras.AssociatedConfigPath, modelExtras.AssociatedCodePath, modelInfo.Token, 0, metadata.FramesCount)
}

func executeJob(videoPath string, modelPath string, configPath string, codePath string, groupID string, startIndex int, framesCount int) {
	command := config.ConfigurationManagerInstance("").DataNodeConfig().IngestionModuleCommand
	args := prepareArgs(videoPath, modelPath, configPath, codePath, groupID, startIndex, framesCount)
	cmd := exec.Command(command, args)
	cmd.Dir = config.ConfigurationManagerInstance("").DataNodeConfig().IngestionModulePath
	err := cmd.Run()
	if errors.IsError(err) {
		log.Println(jobExecutionLoggerPrefix, err)
	}
}

func prepareArgs(videoPath string, modelPath string, configPath string, codePath string, groupID string, startIndex int, frameCount int) string {
	execGroupArg := fmt.Sprintf("execution-group-id=%s", groupID)
	videoPathArg := fmt.Sprintf("video-path=%s", videoPath)
	modelPathArg := fmt.Sprintf("model-path=%s", modelPath)
	configPathArg := fmt.Sprintf("model-config-path=%s", configPath)
	startIndexArg := fmt.Sprintf("start-idx=%d", startIndex)
	frameCountArg := fmt.Sprintf("frame-count=%d", frameCount)
	runCommand := "run"
	args := []string{execGroupArg, videoPathArg, modelPathArg, configPathArg, startIndexArg, frameCountArg, runCommand}
	return strings.Join(args, " ")
}
