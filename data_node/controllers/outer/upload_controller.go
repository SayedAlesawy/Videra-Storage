package outer

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"sync"

	"github.com/SayedAlesawy/Videra-Storage/config"
	datanode "github.com/SayedAlesawy/Videra-Storage/data_node"
	"github.com/SayedAlesawy/Videra-Storage/utils/errors"
	"github.com/julienschmidt/httprouter"
)

var ucLogPrefix = "[Upload-Controller]"

// UploadControllerData represents storage to keep files info
// and keeps track of what files are currently in data node
type UploadControllerData struct {
	fileBase      map[string]datanode.FileInfo // Holds information about files available in data node
	fileBaseMutex sync.RWMutex                 // For safe concurrent access to filebase
	maxChunkSize  int64                        // Maximum acceptable size of received chunk
}

func newUploadControllerData() *UploadControllerData {
	dataNodeConfig := config.ConfigurationManagerInstance("").DataNodeConfig()

	return &UploadControllerData{
		fileBase:     make(map[string]datanode.FileInfo),
		maxChunkSize: dataNodeConfig.MaxRequestSize,
	}
}

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

	w.Header().Set("ID", id)
	w.Header().Set("Max-Request-Size", fmt.Sprintf("%d", server.ucData.maxChunkSize))
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

	if r.ContentLength > server.ucData.maxChunkSize {
		log.Println(ucLogPrefix, r.RemoteAddr, "Request body too large")
		handleRequestError(w, http.StatusBadRequest, fmt.Sprintf("Maximum allowed content length is %d", server.ucData.maxChunkSize))
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, server.ucData.maxChunkSize)

	expectedHeaders := []string{"Offset", "ID"}
	err := validateUploadHeaders(&r.Header, expectedHeaders...)
	if err != nil {
		log.Println(ucLogPrefix, r.RemoteAddr, err)
		handleRequestError(w, http.StatusBadRequest, err.Error())
		return
	}

	id := r.Header.Get("ID")
	if !server.validateIDExistance(id) {
		log.Println(ucLogPrefix, r.RemoteAddr, "ID not found")
		handleRequestError(w, http.StatusForbidden, "ID not found")
		return
	}

	contentLength := r.ContentLength
	offset, err := strconv.ParseInt(r.Header.Get("Offset"), 10, 64)
	if errors.IsError(err) || !server.validateFileOffset(id, offset, contentLength) {
		log.Println(ucLogPrefix, r.RemoteAddr, "Invalid file offset", r.Header.Get("Offset"))
		w.Header().Set("Offset", fmt.Sprintf("%d", server.ucData.fileBase[id].Offset))
		handleRequestError(w, http.StatusBadRequest, "Invalid offset")
		return
	}

	fileInfo := server.ucData.fileBase[id]
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
	file.WriteAt(body, offset)

	server.ucData.fileBaseMutex.Lock()
	defer server.ucData.fileBaseMutex.Unlock()
	log.Println(ucLogPrefix, r.RemoteAddr, filePath, "Writing at offset", fileInfo.Offset)

	fileInfo.Offset += contentLength
	if fileInfo.Offset == fileInfo.Size {
		log.Println(ucLogPrefix, r.RemoteAddr, fmt.Sprintf("File %s was uploaded successfully!", filePath))
		fileInfo.IsCompleted = true

		// Name node should be notified here

		w.WriteHeader(http.StatusCreated)
	}
	server.ucData.fileBase[id] = fileInfo

}

// addNewFile is a function to add new file to storage and file base
func (server *Server) addNewFile(id string, filepath string, filename string, filesize int64) error {
	server.ucData.fileBaseMutex.Lock()
	defer server.ucData.fileBaseMutex.Unlock()

	err := datanode.CreateFile(filepath)
	if errors.IsError(err) {
		return err
	}

	server.ucData.fileBase[id] = datanode.FileInfo{
		Name:        filename,
		Path:        filepath,
		Offset:      0,
		Size:        filesize,
		IsCompleted: false,
	}

	return nil
}

func (server *Server) validateIDExistance(id string) bool {
	server.ucData.fileBaseMutex.RLock()
	defer server.ucData.fileBaseMutex.RUnlock()

	_, present := server.ucData.fileBase[id]
	return present
}

func (server *Server) validateFileOffset(id string, offset int64, chunkSize int64) bool {
	server.ucData.fileBaseMutex.RLock()
	defer server.ucData.fileBaseMutex.RUnlock()

	if offset < 0 {
		return false
	}

	file := server.ucData.fileBase[id]
	if file.Offset == offset && !file.IsCompleted && file.Offset+chunkSize <= file.Size {
		return true
	}
	return false
}
