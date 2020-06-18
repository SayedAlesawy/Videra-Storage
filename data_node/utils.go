package datanode

import (
	"crypto/rand"
	"fmt"
	"os"

	"github.com/SayedAlesawy/Videra-Storage/utils/errors"
)

// CreateFileDirectory creates a directory with given permission
func CreateFileDirectory(dirPath string, perm os.FileMode) error {
	err := os.MkdirAll(dirPath, perm)
	return err
}

// CreateFile A function to create a file given path
func CreateFile(filepath string) error {
	file, err := os.Create(filepath)

	if errors.IsError(err) {
		return err
	}

	defer file.Close()

	return nil
}

// GenerateRandomString generates random string with length 2*n
func GenerateRandomString(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// getNameNodeAddress A function to get the name node address
func (dataNode *DataNode) getNameNodeAddress() string {
	return fmt.Sprintf("%s:%s", dataNode.NameNode.IP, dataNode.NameNode.Port)
}
