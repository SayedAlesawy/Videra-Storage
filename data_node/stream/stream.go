package stream

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/SayedAlesawy/Videra-Storage/config"
	datanode "github.com/SayedAlesawy/Videra-Storage/data_node"
)

var streamEncodingLoggerPrefix = "[Stream-Encoding]"

// PrepareStreamingVideo transforms video into streamable format
func PrepareStreamingVideo(videoInfo datanode.File) {
	config := config.ConfigurationManagerInstance("").DataNodeConfig()

	wd, _ := os.Getwd()
	folderPath := path.Join(wd, config.StreamFolderName, videoInfo.Parent)
	datanode.CreateFileDirectory(folderPath, 0744)

	err := encodeHLS(videoInfo.Path, folderPath)
	if err != nil {
		log.Println(streamEncodingLoggerPrefix, err)
		return
	}

	// removed the working directory part to support streaming server
	folderPath = path.Join(config.StreamFolderName, videoInfo.Parent)
	streamFilePath := path.Join(folderPath, config.StreamPlaylistName)
	err = updateDB(streamFilePath, videoInfo)
	if err != nil {
		log.Println(streamEncodingLoggerPrefix, err)
		return
	}
	log.Println(streamEncodingLoggerPrefix, "Stream file encoded at ", streamFilePath)
}

func encodeHLS(inputFile string, outputFolder string) error {
	command := "ffmpeg"
	args := prepareArgs(inputFile, outputFolder)
	cmd := exec.Command(command, args...)

	err := cmd.Start()
	if err != nil {
		return err
	}

	log.Println(streamEncodingLoggerPrefix, "starting encoding HLS at process", cmd.Process.Pid)
	cmd.Wait()
	return nil
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

func updateDB(streamFilePath string, videoInfo datanode.File) error {
	dn := datanode.NodeInstance()
	return dn.DB.Connection.Model(&videoInfo).Update("HLSPath", streamFilePath).Error
}
