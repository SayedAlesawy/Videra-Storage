package requests

import (
	"fmt"
	"net/http"

	"github.com/SayedAlesawy/Videra-Storage/utils/errors"
)

// TimeStampLayout Layout for timestamp fields
var TimeStampLayout = "2006-01-02T15:04:05.000Z"

// HandleRequestError A function to handle http request failure
func HandleRequestError(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	fmt.Fprintf(w, FormatMessage("error", message))
}

// ValidateHeaders is a function to check existance of parameters inside header
func ValidateHeaders(h *http.Header, params ...string) error {
	for _, param := range params {
		if h.Get(param) == "" {
			return errors.New(fmt.Sprintf("%s header not provided", param))
		}
	}

	return nil
}

// FormatMessage A function to format a message into json
func FormatMessage(key string, message string) string {
	return fmt.Sprintf("{\"%s\": \"%s\"}", key, message)
}
