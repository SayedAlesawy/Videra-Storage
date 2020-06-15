package errors

import (
	"errors"
	"log"
)

// New Creates a custom error with specified msg
func New(msg string) error {
	return errors.New(msg)
}

// IsError A function to check if an error has occurred
func IsError(err error) bool {
	return err != nil
}

// HandleError A function for handling errors
func HandleError(err error, msg string, terminate bool) {
	if err != nil {
		log.Println(msg)

		if terminate {
			log.Panic(err)
		}
	}
}
