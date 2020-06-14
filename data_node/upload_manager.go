package datanode

import (
	"fmt"
	"log"
	"net/http"
	"path"
	"strconv"
	"sync"

	"github.com/SayedAlesawy/Videra-Ingestion/orchestrator/utils/errors"
	"github.com/SayedAlesawy/Videra-Storage/config"
	"github.com/julienschmidt/httprouter"
)

// uploadManagerOnce Used to garauntee thread safety for singleton instances
var uploadManagerOnce sync.Once

// uploadManagerInstance A singleton instance of the upload manager object
var uploadManagerInstance *UploadManager

// UploadManagerInstance A function to return a singleton upload manager instance
func UploadManagerInstance() *UploadManager {

	uploadManagerOnce.Do(func() {
		uploadManager := UploadManager{
			fileBase:  make(map[string]FileInfo),
			logPrefix: "[Upload-Manager]",
		}

		uploadManagerInstance = &uploadManager
	})

	return uploadManagerInstance
}

// Start A function to start listening
func (um *UploadManager) Start() {
	dataNodeConfig := config.ConfigurationManagerInstance("").DataNodeConfig()

	router := httprouter.New()
	router.POST("/upload", um.handleUpload)

	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%s", dataNodeConfig.IP, dataNodeConfig.Port), router))
}

// HandleUpload is upload endpoint handler
func (um *UploadManager) handleUpload(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	log.Println(um.logPrefix, fmt.Sprintf("Received request from %s", r.RemoteAddr))
	reqType := r.Header.Get("Request-Type")

	switch reqType {
	case "INIT":
		um.handleInitialUpload(w, r)
	case "APPEND":

	default:
		log.Println(um.logPrefix, r.RemoteAddr, fmt.Sprintf("request-type header value undefined - %s", reqType))
		handleRequestError(w, http.StatusBadRequest, "Request-Type header value is not undefined")
	}
}

// handleInitialUpload is a function responsible for handling the first upload request
func (um *UploadManager) handleInitialUpload(w http.ResponseWriter, r *http.Request) {
	log.Println(um.logPrefix, r.RemoteAddr, "Received INIT request")

	expectedHeaders := []string{"Content-Length", "Filename"}
	err := um.validateUploadHeaders(&r.Header, expectedHeaders...)

	if err != nil {
		log.Println(um.logPrefix, r.RemoteAddr, err)
		handleRequestError(w, http.StatusBadRequest, err.Error())
		return
	}

	filesize, err := strconv.ParseInt(r.Header.Get("Content-Length"), 10, 64)
	if errors.IsError(err) || filesize <= 0 {
		log.Println(um.logPrefix, r.RemoteAddr, "Error parsing file size")
		handleRequestError(w, http.StatusBadRequest, "Invalid file size")
		return
	}

	id := generateRandomString(10)
	filepath := id
	filename := r.Header.Get("Filename") // Maybe be changed later

	log.Println(um.logPrefix, r.RemoteAddr, "creating file with id", id)
	err = createFileDirectory(filepath, 0744)
	if errors.IsError(err) {
		log.Println(um.logPrefix, r.RemoteAddr, err)
		handleRequestError(w, http.StatusBadRequest, "Internal server error")
		return
	}

	err = um.addNewFile(id, filepath, filename, filesize)
	if errors.IsError(err) {
		log.Println(um.logPrefix, r.RemoteAddr, err)
		handleRequestError(w, http.StatusBadRequest, "Internal server error")
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("ID", id)
}

// validateUploadHeaders is a function to check existance of parameters inside header
func (um *UploadManager) validateUploadHeaders(h *http.Header, params ...string) error {
	for _, param := range params {
		if h.Get(param) == "" {
			return errors.New(fmt.Sprintf("%s header not provided", param))
		}
	}

	return nil
}

// addNewFile is a function to add new file to storage and file base
func (um *UploadManager) addNewFile(id string, filepath string, filename string, filesize int64) error {
	um.fileBaseMutex.Lock()
	defer um.fileBaseMutex.Unlock()

	err := createFile(path.Join(filepath, filename))
	if errors.IsError(err) {
		return err
	}

	um.fileBase[id] = FileInfo{
		Name:        filename,
		Path:        filepath,
		Offset:      0,
		Size:        filesize,
		isCompleted: false,
	}

	return nil
}
