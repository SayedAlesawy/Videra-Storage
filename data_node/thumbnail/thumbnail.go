package thumbnail

import (
	"fmt"
	"log"
	"os/exec"
	"path"
	"strings"

	"github.com/SayedAlesawy/Videra-Storage/config"
	datanode "github.com/SayedAlesawy/Videra-Storage/data_node"
)

var thumbnailLoggerPrefix = "[Thumbnail]"

// PrepareThumbnail generates thumbnail from video
func PrepareThumbnail(videoInfo datanode.File) {
	config := config.ConfigurationManagerInstance("").DataNodeConfig()
	thumbnailFolder := config.ThumbnailFolderName
	datanode.CreateFileDirectory(thumbnailFolder, 0744)

	outputFilePath := path.Join(thumbnailFolder, fmt.Sprintf("%s_thumbnail.jpg", videoInfo.Parent))
	err := generateThumbnail(videoInfo.Path, outputFilePath)
	if err != nil {
		log.Println(thumbnailLoggerPrefix, err)
		return
	}

	err = updateDB(outputFilePath, videoInfo)
	if err != nil {
		log.Println(thumbnailLoggerPrefix, err)
		return
	}
	log.Println(thumbnailLoggerPrefix, "Stream file encoded at ", outputFilePath)
}

func generateThumbnail(inputFile string, outputFilename string) error {
	command := "ffmpeg"
	args := prepareArgs(inputFile, outputFilename)
	cmd := exec.Command(command, args...)

	err := cmd.Start()
	if err != nil {
		return err
	}

	log.Println(thumbnailLoggerPrefix, "starting thumbnail generation at process", cmd.Process.Pid)
	cmd.Wait()
	return nil
}

func prepareArgs(inputFile string, outputFilename string) []string {
	config := config.ConfigurationManagerInstance("").DataNodeConfig()
	args := fmt.Sprintf("-i %s -vframes 1 -an", inputFile)
	args = fmt.Sprintf("%s -s %dx%d", args, config.ThumbnailOutputWidth, config.ThumbnailOutputHeight)
	args = fmt.Sprintf("%s -ss %d", args, config.ThumbnailCaptureSecond)
	args = fmt.Sprintf("%s %s", args, outputFilename)

	return strings.Split(args, " ")
}

func updateDB(thumbnailFilePath string, videoInfo datanode.File) error {
	dn := datanode.NodeInstance()
	return dn.DB.Connection.Model(&videoInfo).Update("ThumbnailPath", thumbnailFilePath).Error
}
