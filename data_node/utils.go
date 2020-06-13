package datanode

import (
	"crypto/rand"
	"fmt"
	"os"
)

//Here goes any utils that are specific to data node

// createFileDirectory creates a directory with given permission
func createFileDirectory(dirPath string, perm os.FileMode) error {
	err := os.MkdirAll(dirPath, perm)
	return err
}

// generateRandomString generates random string with length n
func generateRandomString(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
