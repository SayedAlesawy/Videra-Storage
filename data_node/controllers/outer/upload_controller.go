package outer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/SayedAlesawy/Videra-Storage/config"
	datanode "github.com/SayedAlesawy/Videra-Storage/data_node"
	"github.com/SayedAlesawy/Videra-Storage/data_node/ingest"
	"github.com/SayedAlesawy/Videra-Storage/data_node/replication"
	"github.com/SayedAlesawy/Videra-Storage/utils/errors"
	"github.com/SayedAlesawy/Videra-Storage/utils/requests"
	"github.com/julienschmidt/httprouter"
)

var ucLogPrefix = "[Upload-Controller]"

//ModelUploadOrder represents the order in which model files will be uploaded
var modelUploadOrder = [...]string{"model", "config", "code"}

// UploadRequestHandler is upload endpoint handler
func (server *Server) UploadRequestHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	reqType := strings.ToLower(r.Header.Get("Request-Type"))

	switch reqType {
	case "init":
		server.handleInitialUpload(w, r)
	case "append":
		server.handleAppendUpload(w, r)
	default:
		log.Println(ucLogPrefix, r.RemoteAddr, fmt.Sprintf("request-type header value undefined - %s", reqType))

		requests.HandleRequestError(w, http.StatusBadRequest, "Request-Type header value is not undefined")
	}
}

// handleInitialUpload is a function responsible for handling the first upload request
func (server *Server) handleInitialUpload(w http.ResponseWriter, r *http.Request) {
	log.Println(ucLogPrefix, r.RemoteAddr, "Received init request")

	expectedHeaders := []string{"Filename", "Filesize", "Filetype"}
	err := requests.ValidateHeaders(&r.Header, expectedHeaders...)

	if err != nil {
		log.Println(ucLogPrefix, r.RemoteAddr, err)
		requests.HandleRequestError(w, http.StatusBadRequest, err.Error())
		return
	}

	fileType := strings.ToLower(r.Header.Get("Filetype"))
	if !server.validateFileTypes(fileType) {
		log.Println(ucLogPrefix, r.RemoteAddr, "Unsupported file type", fileType)
		handleRequestError(w, http.StatusBadRequest, fmt.Sprintf("Supported types are video and model"))
		return
	}

	switch fileType {
	case datanode.ModelFileType:
		server.handleModelInitialUpload(w, r)
	case datanode.VideoFileType:
		server.handleVideoInitialUpload(w, r)
	}
}

// handleModelInitialUpload is responsible for handling upload request for model file
func (server *Server) handleModelInitialUpload(w http.ResponseWriter, r *http.Request) {
	expectedHeaders := []string{"Model-Size", "Config-Size", "Code-Size"}
	err := requests.ValidateHeaders(&r.Header, expectedHeaders...)
	if err != nil {
		log.Println(ucLogPrefix, r.RemoteAddr, err)
		handleRequestError(w, http.StatusBadRequest, err.Error())
		return
	}

	err = server.validateModelSize(&r.Header)
	if errors.IsError(err) {
		log.Println(ucLogPrefix, r.RemoteAddr, err)
		handleRequestError(w, http.StatusBadRequest, err.Error())
		return
	}
	filesize, _ := strconv.ParseInt(r.Header.Get("Filesize"), 10, 64)
	modelSize, _ := strconv.ParseInt(r.Header.Get("Model-Size"), 10, 64)
	configSize, _ := strconv.ParseInt(r.Header.Get("Config-Size"), 10, 64)
	codeSize, _ := strconv.ParseInt(r.Header.Get("Code-Size"), 10, 64)

	id := datanode.GenerateRandomString(10)
	filename := r.Header.Get("Filename")
	fileType := strings.ToLower(r.Header.Get("Filetype"))

	wd, _ := os.Getwd()
	// file will be at path .../files/id/filaname
	folderpath := path.Join(wd, "files", id)

	modelPath := path.Join(folderpath, filename)
	configPath := path.Join(folderpath, fmt.Sprintf("%s_config.conf", id))
	codePath := path.Join(folderpath, "code_file.py")

	log.Println(ucLogPrefix, r.RemoteAddr, "creating file with id", id)
	err = datanode.CreateFileDirectory(folderpath, 0744)
	if errors.IsError(err) {
		log.Println(ucLogPrefix, r.RemoteAddr, err)
		handleRequestError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	err = datanode.CreateFile(modelPath)
	if errors.IsError(err) {
		log.Println(ucLogPrefix, r.RemoteAddr, err)
		handleRequestError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	err = datanode.CreateFile(configPath)
	if errors.IsError(err) {
		log.Println(ucLogPrefix, r.RemoteAddr, err)
		handleRequestError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	err = datanode.CreateFile(codePath)
	if errors.IsError(err) {
		log.Println(ucLogPrefix, r.RemoteAddr, err)
		handleRequestError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	extras := datanode.ModelExtras{
		ModelSize:            modelSize,
		AssociatedConfigPath: configPath,
		AssociatedConfigSize: configSize,
		AssociatedCodePath:   codePath,
		AssociatedCodeSize:   codeSize,
	}

	parentID := id
	if r.Header.Get("Parent") != "" {
		parentID = r.Header.Get("Parent")
	}

	extrasBytes, _ := json.Marshal(extras)
	//Insert a file info record in the database
	err = datanode.NodeInstance().DB.Connection.Create(&datanode.File{
		Token:      id,
		Name:       filename,
		Type:       fileType,
		Path:       modelPath,
		Extras:     string(extrasBytes),
		Size:       filesize,
		DataNodeID: datanode.NodeInstance().ID,
		Parent:     parentID,
		Offset:     0,
	}).Error
	if errors.IsError(err) {
		log.Println(ucLogPrefix, r.RemoteAddr, err)
		handleRequestError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	maxRequestSize := config.ConfigurationManagerInstance("").DataNodeConfig().MaxRequestSize

	w.Header().Set("ID", id)
	w.Header().Set("Max-Request-Size", fmt.Sprintf("%d", maxRequestSize))
	uploadOrderJSON, _ := json.Marshal(modelUploadOrder)
	w.Header().Set("Upload-Order", string(uploadOrderJSON))
	w.WriteHeader(http.StatusCreated)
}

// validateModelSize validates the model and associated files size
func (server *Server) validateModelSize(h *http.Header) error {
	requiredSizes := []string{h.Get("Filesize"), h.Get("Model-Size"), h.Get("Config-Size"), h.Get("Code-Size")}
	for _, req := range requiredSizes {
		err := isValideSize(req)
		if errors.IsError(err) {
			return errors.New(fmt.Sprintf("Invalid %s", req))
		}
	}

	// in model case, filesize should be the sum of (model,config,code) sizes
	filesize, _ := strconv.ParseInt(h.Get("Filesize"), 10, 64)
	modelSize, _ := strconv.ParseInt(h.Get("Model-Size"), 10, 64)
	configSize, _ := strconv.ParseInt(h.Get("Config-Size"), 10, 64)
	codeSize, _ := strconv.ParseInt(h.Get("Code-Size"), 10, 64)
	if filesize != modelSize+configSize+codeSize {
		return errors.New("Invalid filesize")
	}
	return nil
}

// handleVideoInitialUpload is responsible for handling upload request for video file
func (server *Server) handleVideoInitialUpload(w http.ResponseWriter, r *http.Request) {
	associatedModelID := r.Header.Get("Associated-Model-ID")
	if associatedModelID == "" {
		log.Println(ucLogPrefix, r.RemoteAddr, "Associated Model ID not provided")
		requests.HandleRequestError(w, http.StatusBadRequest, "Associated Model ID not provided")
		return
	}
	notFound := datanode.NodeInstance().DB.Connection.Where("parent = ?", associatedModelID).RecordNotFound()
	if notFound {
		log.Println(ucLogPrefix, r.RemoteAddr, fmt.Sprintf("Model with token: %s is not found", associatedModelID))
		requests.HandleRequestError(w, http.StatusNotFound, fmt.Sprintf("Record with token: %s is not found", associatedModelID))
		return
	}
	err := isValideSize(r.Header.Get("Filesize"))
	if errors.IsError(err) {
		log.Println(ucLogPrefix, r.RemoteAddr, "Error parsing file size")
		requests.HandleRequestError(w, http.StatusBadRequest, "Invalid file size")
		return
	}

	filesize, _ := strconv.ParseInt(r.Header.Get("Filesize"), 10, 64)
	id := datanode.GenerateRandomString(10)
	filename := r.Header.Get("Filename")
	fileType := strings.ToLower(r.Header.Get("Filetype"))
	wd, _ := os.Getwd()
	// file will be at path .../files/id/filaname
	folderpath := path.Join(wd, "files", id)
	filepath := path.Join(folderpath, filename)

	log.Println(ucLogPrefix, r.RemoteAddr, "creating file with id", id)
	err = datanode.CreateFileDirectory(folderpath, 0744)
	if errors.IsError(err) {
		log.Println(ucLogPrefix, r.RemoteAddr, err)
		requests.HandleRequestError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	err = datanode.CreateFile(filepath)
	if errors.IsError(err) {
		log.Println(ucLogPrefix, r.RemoteAddr, err)
		requests.HandleRequestError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	metadata := datanode.VideoMetadata{}
	metadata.AssociatedModel = associatedModelID
	metadataJSON, err := json.Marshal(metadata)
	if errors.IsError(err) {
		log.Println(ucLogPrefix, r.RemoteAddr, err)
		requests.HandleRequestError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	parentID := id
	if r.Header.Get("Parent") != "" {
		parentID = r.Header.Get("Parent")
	}

	// original node
	if parentID == id {
		log.Println(ucLogPrefix, "Send replicated init")
		err := replication.ReplicateVideo(r, id)
		if errors.IsError(err) {
			log.Println(ucLogPrefix, r.RemoteAddr, err)
			requests.HandleRequestError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
	}
	//Insert a file info record in the database
	err = datanode.NodeInstance().DB.Connection.Create(&datanode.File{
		Token:      id,
		Name:       filename,
		Type:       fileType,
		Path:       filepath,
		Size:       filesize,
		Extras:     string(metadataJSON),
		DataNodeID: datanode.NodeInstance().ID,
		Parent:     parentID,
		Offset:     0,
	}).Error
	if errors.IsError(err) {
		log.Println(ucLogPrefix, r.RemoteAddr, err)
		requests.HandleRequestError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	maxRequestSize := config.ConfigurationManagerInstance("").DataNodeConfig().MaxRequestSize

	w.Header().Set("ID", id)
	w.Header().Set("Max-Request-Size", fmt.Sprintf("%d", maxRequestSize))
	w.WriteHeader(http.StatusCreated)
}

// handleAppendUpload is a function responsible for handling the first upload request
func (server *Server) handleAppendUpload(w http.ResponseWriter, r *http.Request) {
	log.Println(ucLogPrefix, r.RemoteAddr, "Received APPEND request")
	// Content length not provided
	if r.ContentLength <= 0 {
		log.Println(ucLogPrefix, r.RemoteAddr, "Content-Length header not provided")
		requests.HandleRequestError(w, http.StatusBadRequest, "Content-Length header not provided")
		return
	}

	maxRequestSize := config.ConfigurationManagerInstance("").DataNodeConfig().MaxRequestSize
	if r.ContentLength > maxRequestSize {
		log.Println(ucLogPrefix, r.RemoteAddr, "Request body too large", r.ContentLength)
		w.Header().Set("Max-Request-Size", fmt.Sprintf("%v", maxRequestSize))
		handleRequestError(w, http.StatusBadRequest, fmt.Sprintf("Maximum allowed content length is %d", maxRequestSize))
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestSize)

	expectedHeaders := []string{"Offset", "ID"}
	err := requests.ValidateHeaders(&r.Header, expectedHeaders...)
	if err != nil {
		log.Println(ucLogPrefix, r.RemoteAddr, err)
		requests.HandleRequestError(w, http.StatusBadRequest, err.Error())
		return
	}

	//Fetch the file info from the database
	id := r.Header.Get("ID")
	var fileInfo datanode.File

	notFound := datanode.NodeInstance().DB.Connection.Where("token = ?", id).Find(&fileInfo).RecordNotFound()
	if notFound {
		log.Println(ucLogPrefix, r.RemoteAddr, fmt.Sprintf("Record with token: %s is not found", id))
		requests.HandleRequestError(w, http.StatusNotFound, fmt.Sprintf("Record with token: %s is not found", id))
		return
	}

	if server.isFileComplete(fileInfo) {
		log.Println(ucLogPrefix, r.RemoteAddr, "File was completed from previous upload")
		w.WriteHeader(http.StatusCreated)
		return
	}

	contentLength := r.ContentLength
	offset, err := strconv.ParseInt(r.Header.Get("Offset"), 10, 64)
	if errors.IsError(err) || !server.validateFileOffset(fileInfo, offset, contentLength) {
		log.Println(ucLogPrefix, r.RemoteAddr, fmt.Sprintf("Invalid file offset, expected %v found %v", fileInfo.Offset, r.Header.Get("Offset")))
		w.Header().Set("Offset", fmt.Sprintf("%d", fileInfo.Offset))
		requests.HandleRequestError(w, http.StatusBadRequest, "Invalid offset")
		return
	}

	filePath := fileInfo.Path

	// writeoffset is offset at file which will be written at,
	// in video file it's the same as offset variable,
	// but in model it's relative based on which file is being written at,
	// for example offset maybe 500 but it will write at offset 0 of config file
	var writeOffset int64
	writeOffset = offset

	if fileInfo.Type == datanode.ModelFileType {
		var modelExtras datanode.ModelExtras
		err := json.Unmarshal([]byte(fileInfo.Extras), &modelExtras)
		if errors.IsError(err) {
			log.Println(ucLogPrefix, r.RemoteAddr, err)
			handleRequestError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		// ****Model****/**Config**/*Code*/
		if offset >= modelExtras.ModelSize {
			// chunk is either belongs to config file or code file
			if offset < modelExtras.ModelSize+modelExtras.AssociatedConfigSize {
				filePath = modelExtras.AssociatedConfigPath
				writeOffset -= modelExtras.ModelSize
			} else if offset >= modelExtras.ModelSize+modelExtras.AssociatedConfigSize {
				filePath = modelExtras.AssociatedCodePath
				writeOffset -= modelExtras.ModelSize + modelExtras.AssociatedConfigSize
			}
		}
	}

	file, err := os.OpenFile(filePath, os.O_WRONLY, 0644)
	defer file.Close()
	if errors.IsError(err) {
		log.Println(ucLogPrefix, r.RemoteAddr, err)
		requests.HandleRequestError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if errors.IsError(err) {
		log.Println(ucLogPrefix, r.RemoteAddr, err)
		requests.HandleRequestError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	log.Println(ucLogPrefix, r.RemoteAddr, filePath, "Writing at offset", fileInfo.Offset)
	file.WriteAt(body, writeOffset)

	//Update values
	fileInfo.Offset += contentLength
	if fileInfo.Offset == fileInfo.Size {
		now := time.Now()
		fileInfo.CompletedAt = &now

		if fileInfo.Type == datanode.VideoFileType {
			var videoMetadata datanode.VideoMetadata
			json.Unmarshal([]byte(fileInfo.Extras), &videoMetadata)
			err := server.fetchVideoMetaData(fileInfo, &videoMetadata)
			if errors.IsError(err) {
				log.Println(ucLogPrefix, r.RemoteAddr, err)
				handleRequestError(w, http.StatusInternalServerError, "Internal server error")
				return
			}
			metaData, err := json.Marshal(videoMetadata)
			fileInfo.Extras = string(metaData)
		}
	}

	if fileInfo.Type == datanode.VideoFileType && !isReplica(fileInfo) {
		r.Body = ioutil.NopCloser(bytes.NewReader(body))
		err := replication.ReplicateVideo(r, id)
		if errors.IsError(err) {
			log.Println(ucLogPrefix, r.RemoteAddr, err)
			handleRequestError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
	}

	err = datanode.NodeInstance().DB.Connection.Save(&fileInfo).Error
	if errors.IsError(err) {
		log.Println(ucLogPrefix, r.RemoteAddr, err)
		requests.HandleRequestError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if fileInfo.Offset == fileInfo.Size {
		if fileInfo.Type == datanode.VideoFileType {
			if !isReplica(fileInfo) {
				go ingest.StartJob(fileInfo)
			} else {
				log.Println("Replication done")
			}
		}

		log.Println(ucLogPrefix, r.RemoteAddr, fmt.Sprintf("File %s was uploaded successfully!", filePath))
		w.WriteHeader(http.StatusCreated)
	}

}

// validateFileOffset A function validate file offset
func (server *Server) validateFileOffset(fileinfo datanode.File, offset int64, chunkSize int64) bool {
	if offset < 0 {
		return false
	}

	if fileinfo.Offset == offset && fileinfo.Offset+chunkSize <= fileinfo.Size {
		return true
	}

	return false
}

// isFileComplete A function to check if file upload was completed previously
func (server *Server) isFileComplete(fileinfo datanode.File) bool {
	return fileinfo.CompletedAt != nil
}

func (server *Server) validateFileTypes(fileType string) bool {
	for _, supportedFileType := range datanode.SupportedFileTypes {
		if supportedFileType == fileType {
			return true
		}
	}
	return true
}

// fetchVideoMetaData is a function responsible of retrieving metadata from video file
func (server *Server) fetchVideoMetaData(fileInfo datanode.File, extras *datanode.VideoMetadata) error {
	dataNodeConfig := config.ConfigurationManagerInstance("").DataNodeConfig()

	//temp file to save metadata
	tempFile, err := ioutil.TempFile("", "*_metadata.txt")
	if err != nil {
		return err
	}
	tempFile.Close()
	defer os.Remove(tempFile.Name()) //remove the temp file at the end

	command := dataNodeConfig.MetadataCommand
	script := dataNodeConfig.MetadataScriptPath
	inputFile := fileInfo.Path
	outputFile := tempFile.Name()

	cmd := exec.Command(command, script, "-i", inputFile, "-o", outputFile)
	err = cmd.Run()
	if errors.IsError(err) {
		return err
	}
	cmd.Wait()

	content, err := ioutil.ReadFile(outputFile)
	if errors.IsError(err) {
		return err
	}
	err = json.Unmarshal(content, extras)
	if errors.IsError(err) {
		return err
	}
	if extras.FramesCount == 0 {
		return errors.New("Error parsing file")
	}
	return nil
}
