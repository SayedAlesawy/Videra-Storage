package outer

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/SayedAlesawy/Videra-Storage/config"
	datanode "github.com/SayedAlesawy/Videra-Storage/data_node"
	"github.com/SayedAlesawy/Videra-Storage/utils/errors"
	"github.com/julienschmidt/httprouter"
)

var ucLogPrefix = "[Upload-Controller]"

// UploadRequestHandler is upload endpoint handler
func (server *Server) UploadRequestHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	reqType := r.Header.Get("Request-Type")

	switch reqType {
	case "INIT":
		server.handleInitialUpload(w, r)
	case "APPEND":
		server.handleAppendUpload(w, r)
	default:
		log.Println(ucLogPrefix, r.RemoteAddr, fmt.Sprintf("request-type header value undefined - %s", reqType))

		handleRequestError(w, http.StatusBadRequest, "Request-Type header value is not undefined")
	}
}

// handleInitialUpload is a function responsible for handling the first upload request
func (server *Server) handleInitialUpload(w http.ResponseWriter, r *http.Request) {
	log.Println(ucLogPrefix, r.RemoteAddr, "Received INIT request")

	expectedHeaders := []string{"Filename", "Filesize"}
	err := validateUploadHeaders(&r.Header, expectedHeaders...)

	if err != nil {
		log.Println(ucLogPrefix, r.RemoteAddr, err)
		handleRequestError(w, http.StatusBadRequest, err.Error())
		return
	}

	filesize, err := strconv.ParseInt(r.Header.Get("Filesize"), 10, 64)
	if errors.IsError(err) || filesize <= 0 {
		log.Println(ucLogPrefix, r.RemoteAddr, "Error parsing file size")
		handleRequestError(w, http.StatusBadRequest, "Invalid file size")
		return
	}

	id := datanode.GenerateRandomString(10)
	filename := r.Header.Get("Filename") // Maybe be changed later
	wd, _ := os.Getwd()
	// file will be at path .../files/id/filaname
	folderpath := path.Join(wd, "files", id)
	filepath := path.Join(folderpath, filename)
	log.Println(ucLogPrefix, r.RemoteAddr, "creating file with id", id)
	err = datanode.CreateFileDirectory(folderpath, 0744)
	if errors.IsError(err) {
		log.Println(ucLogPrefix, r.RemoteAddr, err)
		handleRequestError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	err = server.addNewFile(id, filepath, filename, filesize)
	if errors.IsError(err) {
		log.Println(ucLogPrefix, r.RemoteAddr, err)
		handleRequestError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	//Insert a file info record in the database
	err = datanode.NodeInstance().DB.Connection.Create(&datanode.File{
		Token:      id,
		Name:       filename,
		Path:       filepath,
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
	w.WriteHeader(http.StatusCreated)
}

// handleAppendUpload is a function responsible for handling the first upload request
func (server *Server) handleAppendUpload(w http.ResponseWriter, r *http.Request) {
	log.Println(ucLogPrefix, r.RemoteAddr, "Received APPEND request")
	// Content length not provided
	if r.ContentLength <= 0 {
		log.Println(ucLogPrefix, r.RemoteAddr, "Content-Length header not provided")
		handleRequestError(w, http.StatusBadRequest, "Content-Length header not provided")
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
	err := validateUploadHeaders(&r.Header, expectedHeaders...)
	if err != nil {
		log.Println(ucLogPrefix, r.RemoteAddr, err)
		handleRequestError(w, http.StatusBadRequest, err.Error())
		return
	}

	//Fetch the file info from the database
	id := r.Header.Get("ID")
	var fileInfo datanode.File

	notFound := datanode.NodeInstance().DB.Connection.Where("token = ?", id).Find(&fileInfo).RecordNotFound()
	if notFound {
		log.Println(ucLogPrefix, r.RemoteAddr, fmt.Sprintf("Record with token: %s is not found", id))
		handleRequestError(w, http.StatusNotFound, fmt.Sprintf("Record with token: %s is not found", id))
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
		handleRequestError(w, http.StatusBadRequest, "Invalid offset")
		return
	}

	filePath := fileInfo.Path
	file, err := os.OpenFile(filePath, os.O_WRONLY, 0644)
	defer file.Close()
	if errors.IsError(err) {
		log.Println(ucLogPrefix, r.RemoteAddr, err)
		handleRequestError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if errors.IsError(err) {
		log.Println(ucLogPrefix, r.RemoteAddr, err)
		handleRequestError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	log.Println(ucLogPrefix, r.RemoteAddr, filePath, "Writing at offset", fileInfo.Offset)
	file.WriteAt(body, offset)

	//Update values
	fileInfo.Offset += contentLength
	if fileInfo.Offset == fileInfo.Size {
		now := time.Now()
		fileInfo.CompletedAt = &now
	}

	err = datanode.NodeInstance().DB.Connection.Save(&fileInfo).Error
	if errors.IsError(err) {
		log.Println(ucLogPrefix, r.RemoteAddr, err)
		handleRequestError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if fileInfo.Offset == fileInfo.Size {
		log.Println(ucLogPrefix, r.RemoteAddr, fmt.Sprintf("File %s was uploaded successfully!", filePath))
		w.WriteHeader(http.StatusCreated)
	}
}

// addNewFile is a function to add new file to storage and file base
func (server *Server) addNewFile(id string, filepath string, filename string, filesize int64) error {
	err := datanode.CreateFile(filepath)
	if errors.IsError(err) {
		return err
	}

	return nil
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
