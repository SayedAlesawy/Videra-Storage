package outer

import (
	"encoding/json"
	"log"
	"os/exec"

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

	executeJob(videoInfo.Path, modelInfo.Path, modelExtras.AssociatedConfigPath, modelExtras.AssociatedCodePath)
}

func executeJob(videoPath string, modelPath string, configPath string, codePath string) {
	command := config.ConfigurationManagerInstance("").DataNodeConfig().IngestionModuleCommand
	cmd := exec.Command(command)
	cmd.Dir = config.ConfigurationManagerInstance("").DataNodeConfig().IngestionModulePath
	err := cmd.Run()
	if errors.IsError(err) {
		log.Println(jobExecutionLoggerPrefix, err)
	}
}
