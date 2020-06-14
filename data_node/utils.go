package datanode

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"os"

	"github.com/SayedAlesawy/Videra-Ingestion/orchestrator/utils/errors"
)

//Here goes any utils that are specific to data node

// createFileDirectory creates a directory with given permission
func createFileDirectory(dirPath string, perm os.FileMode) error {
	err := os.MkdirAll(dirPath, perm)
	return err
}

// createFile A function to create a file given path
func createFile(filepath string) error {
	file, err := os.Create(filepath)

	if errors.IsError(err) {
		return err
	}

	defer file.Close()

	return nil
}

// generateRandomString generates random string with length n
func generateRandomString(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// handleRequestError A function to handle http request failure
func handleRequestError(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	w.Write([]byte(message))
}

// getNameNodeAddress A function to get the name node address
func (dataNode *DataNode) getNameNodeAddress() string {
	return fmt.Sprintf("%s:%s", dataNode.NameNode.IP, dataNode.NameNode.Port)
}
