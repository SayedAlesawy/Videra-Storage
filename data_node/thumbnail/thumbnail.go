package thumbnail

import (
	"fmt"
	"log"
	"path"
	"strings"

	"github.com/SayedAlesawy/Videra-Storage/config"
	datanode "github.com/SayedAlesawy/Videra-Storage/data_node"
	jobscheduler "github.com/SayedAlesawy/Videra-Storage/data_node/jobs_scheduler"
)

var thumbnailLoggerPrefix = "[Thumbnail]"

// PrepareThumbnail generates thumbnail from video
func PrepareThumbnail(videoInfo datanode.File) {
	config := config.ConfigurationManagerInstance("").DataNodeConfig()
	thumbnailFolder := config.ThumbnailFolderName
	datanode.CreateFileDirectory(thumbnailFolder, 0744)

	outputFilePath := path.Join(thumbnailFolder, fmt.Sprintf("%s_thumbnail.jpg", videoInfo.Parent))

	command := "ffmpeg"
	args := prepareArgs(videoInfo.Path, outputFilePath)
	name := getJobName(videoInfo.Parent)
	tableName := datanode.NodeInstance().DB.Connection.NewScope(videoInfo).TableName()
	columnName := "thumbnail_path"
	post := jobscheduler.PostJob{ID: videoInfo.ID, TableName: tableName, ColumnName: columnName, NewValue: outputFilePath}

	jobScheduler := jobscheduler.JobQueueInstance()
	jobScheduler.InsertJob(name, command, args, post)
	log.Println(thumbnailLoggerPrefix, "Submitted thumbnail generation for file", videoInfo.Parent)
}

func prepareArgs(inputFile string, outputFilename string) []string {
	config := config.ConfigurationManagerInstance("").DataNodeConfig()
	args := fmt.Sprintf("-i %s -vframes 1 -an", inputFile)
	args = fmt.Sprintf("%s -s %dx%d", args, config.ThumbnailOutputWidth, config.ThumbnailOutputHeight)
	args = fmt.Sprintf("%s -ss %d", args, config.ThumbnailCaptureSecond)
	args = fmt.Sprintf("%s %s", args, outputFilename)

	return strings.Split(args, " ")
}

func getJobName(token string) string {
	return fmt.Sprintf("Thumbnail-%s", token)
}
