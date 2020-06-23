package outer

import (
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
	err := validateUploadHeaders(&r.Header, expectedHeaders...)

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
	err := validateUploadHeaders(&r.Header, expectedHeaders...)
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
	codePath := path.Join(folderpath, fmt.Sprintf("%s_code.py", id))

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

	//Insert a file info record in the database
	err = datanode.NodeInstance().DB.Connection.Create(&datanode.File{
		Token:      id,
		Name:       filename,
		Type:       fileType,
		Path:       filepath,
		Size:       filesize,
		DataNodeID: datanode.NodeInstance().ID,
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
	log.Println(ucLogPrefix, r.RemoteAddr, "Received append request")
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
	err := requests.ValidateUploadHeaders(&r.Header, expectedHeaders...)
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

	if fileInfo.Type == datanode.ConfigFileType {
		err := validateUploadHeaders(&r.Header, "Associated-Model-ID")
		if err != nil {
			log.Println(ucLogPrefix, r.RemoteAddr, err)
			handleRequestError(w, http.StatusBadRequest, err.Error())
			return
		}
		associatedModelID := r.Header.Get("Associated-Model-ID")
		err = server.validateAssociatedModel(associatedModelID)
		if err != nil {
			log.Println(ucLogPrefix, r.RemoteAddr, err.Error())
			handleRequestError(w, http.StatusNotFound, err.Error())
			return
		}
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
	file.WriteAt(body, offset)

	//Update values
	fileInfo.Offset += contentLength
	if fileInfo.Offset == fileInfo.Size && fileInfo.Type != datanode.ModelFileType {
		now := time.Now()
		fileInfo.CompletedAt = &now

		if fileInfo.Type == datanode.ConfigFileType {
			// Update associated file info
			associatedModelID := r.Header.Get("Associated-Model-ID")
			modelFileInfo, err := server.updateAssociatedModel(associatedModelID, fileInfo)
			if errors.IsError(err) {
				log.Println(ucLogPrefix, r.RemoteAddr, err)
				handleRequestError(w, http.StatusInternalServerError, "Internal server error")
				return
			}
			err = datanode.NodeInstance().DB.Connection.Save(&modelFileInfo).Error
			if errors.IsError(err) {
				log.Println(ucLogPrefix, r.RemoteAddr, err)
				handleRequestError(w, http.StatusInternalServerError, "Internal server error")
				return
			}
		} else if fileInfo.Type == datanode.VideoFileType {
			videoMetadata, err := server.fetchVideoMetaData(fileInfo)
			if errors.IsError(err) {
				log.Println(ucLogPrefix, r.RemoteAddr, err)
				handleRequestError(w, http.StatusInternalServerError, "Internal server error")
				return
			}
			fileInfo.Extras = videoMetadata
		}
	}

	err = datanode.NodeInstance().DB.Connection.Save(&fileInfo).Error
	if errors.IsError(err) {
		log.Println(ucLogPrefix, r.RemoteAddr, err)
		requests.HandleRequestError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if fileInfo.Offset == fileInfo.Size {
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

//validateAssociatedModel is a function responsible for validating existance of model with id parameter
func (server *Server) validateAssociatedModel(associatedModelID string) error {
	var modelFileInfo datanode.File
	modelNotFound := datanode.NodeInstance().DB.Connection.Where("token = ? and type = ?", associatedModelID, datanode.ModelFileType).Find(&modelFileInfo).RecordNotFound()
	if modelNotFound {
		return errors.New("Invalid Model")
	}

	if server.isFileComplete(modelFileInfo) {
		return errors.New("Model was associated with another config")
	}

	if modelFileInfo.Offset != modelFileInfo.Size {
		return errors.New("Model wasn't completly uploaded")
	}

	return nil
}

//updateAssociatedModel is a function responsible model database entry associated with config file
func (server *Server) updateAssociatedModel(associatedModelID string, configInfo datanode.File) (datanode.File, error) {
	var modelFileInfo datanode.File

	err := datanode.NodeInstance().DB.Connection.Where("token = ?", associatedModelID).Find(&modelFileInfo).Error
	if err != nil {
		return modelFileInfo, err
	}

	extras := struct {
		AssociatedConfigID   string `json:"associated_config_ID"`
		AssociatedConfigPath string `json:"associated_config_path"`
	}{
		configInfo.Token,
		configInfo.Path,
	}
	marsh, _ := json.Marshal(extras)

	t := time.Now()
	modelFileInfo.CompletedAt = &t
	modelFileInfo.Extras = string(marsh)
	return modelFileInfo, nil
}

// fetchVideoMetaData is a function responsible of retrieving metadata from video file
func (server *Server) fetchVideoMetaData(fileInfo datanode.File) (string, error) {
	dataNodeConfig := config.ConfigurationManagerInstance("").DataNodeConfig()
	command := dataNodeConfig.MetadataCommand
	script := dataNodeConfig.MetadataScriptPath
	inputFile := fileInfo.Path
	outputFile := fileInfo.Path + fileInfo.Token //just a temp file to save metadata

	fmt.Println(inputFile, outputFile)
	// this should be replaced with data from config file
	cmd := exec.Command(command, script, "-i", inputFile, "-o", outputFile)
	err := cmd.Run()
	if errors.IsError(err) {
		return "", err
	}
	cmd.Wait()
	content, err := ioutil.ReadFile(outputFile)
	if errors.IsError(err) {
		return "", err
	}
	os.Remove(outputFile) //remove the temp file
	text := string(content)
	return text, nil
}
