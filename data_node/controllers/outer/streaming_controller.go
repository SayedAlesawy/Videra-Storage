package outer

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// File server to serve files in directory
var fileServer = http.FileServer(http.Dir("stream"))

// StreamingHandler is a handle responsible for serving streaming requests
func (server *Server) StreamingHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// Some browsers doesn't allow reading files from other directory
	// so we can set this header, or allow CORS in the browser
	// but this header isn't recommended in production
	w.Header().Set("Access-Control-Allow-Origin", "*")
	r.URL.Path = p.ByName("filepath")
	fileServer.ServeHTTP(w, r)
}
