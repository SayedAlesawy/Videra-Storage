package outer

import (
	"fmt"
	"net/http"

	"github.com/SayedAlesawy/Videra-Storage/utils/errors"
)

// handleRequestError A function to handle http request failure
func handleRequestError(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	w.Write([]byte(message))
}

// validateUploadHeaders is a function to check existance of parameters inside header
func validateUploadHeaders(h *http.Header, params ...string) error {
	for _, param := range params {
		if h.Get(param) == "" {
			return errors.New(fmt.Sprintf("%s header not provided", param))
		}
	}

	return nil
}
