package stream

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/SayedAlesawy/Videra-Storage/config"
	datanode "github.com/SayedAlesawy/Videra-Storage/data_node"
	jobscheduler "github.com/SayedAlesawy/Videra-Storage/data_node/jobs_scheduler"
)

var streamEncodingLoggerPrefix = "[Stream-Encoding]"

// PrepareStreamingVideo transforms video into streamable format
func PrepareStreamingVideo(videoInfo datanode.File) {
	config := config.ConfigurationManagerInstance("").DataNodeConfig()

	wd, _ := os.Getwd()
	folderPath := path.Join(wd, config.StreamFolderName, videoInfo.Parent)
	datanode.CreateFileDirectory(folderPath, 0744)

	// removed the working directory part to support streaming server
	folderPath = path.Join(config.StreamFolderName, videoInfo.Parent)
	streamFilePath := path.Join(folderPath, config.StreamPlaylistName)

	command := "ffmpeg"
	args := prepareArgs(videoInfo.Path, folderPath)
	name := getJobName(videoInfo.Parent)
	tableName := datanode.NodeInstance().DB.Connection.NewScope(videoInfo).TableName()
	columnName := "hls_path"
	post := jobscheduler.PostJob{ID: videoInfo.ID, TableName: tableName, ColumnName: columnName, NewValue: streamFilePath}

	jobScheduler := jobscheduler.JobQueueInstance()
	jobScheduler.InsertJob(name, command, args, post)
	log.Println(streamEncodingLoggerPrefix, "Submitted hls encode for file", videoInfo.Parent)
}

func prepareArgs(inputFile string, outputFolder string) []string {
	config := config.ConfigurationManagerInstance("").DataNodeConfig()
	args := fmt.Sprintf("-i %s", inputFile)
	args = fmt.Sprintf("%s %s", args, "-profile:v baseline -level 3.0")
	args = fmt.Sprintf("%s -s %dx%d -start_number 0", args, config.StreamOutputVideoWidth, config.StreamOutputVideoHeight)
	args = fmt.Sprintf("%s -hls_time %d", args, config.StreamSegmentTime)
	args = fmt.Sprintf("%s -hls_list_size 0 -f hls %s/%s", args, outputFolder, config.StreamPlaylistName)

	return strings.Split(args, " ")
}

func getJobName(token string) string {
	return fmt.Sprintf("HLS-%s", token)
}
