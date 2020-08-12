package outer

import (
	"net/http"
	"sync"

	"github.com/SayedAlesawy/Videra-Storage/config"
	"github.com/julienschmidt/httprouter"
)

// streamFileServerOnce Used to garauntee thread safety for singleton instances
var streamFileServerOnce sync.Once

// streamFileServerInstance A singleton instance of the streamFileServer object, to serve files in directory
var streamFileServer http.Handler

// streamFileServerInstance A function to return a singleton streamFileServer instance
func streamFileServerInstance() http.Handler {
	dataNodeConfig := config.ConfigurationManagerInstance("").DataNodeConfig()

	streamFileServerOnce.Do(func() {
		streamFileServer = http.FileServer(http.Dir(dataNodeConfig.StreamFolderName))
	})

	return streamFileServer
}

// StreamingHandler is a handle responsible for serving streaming requests
func (server *Server) StreamingHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// Some browsers doesn't allow reading files from other directory
	// so we can set this header, or allow CORS in the browser
	// but this header isn't recommended in production
	w.Header().Set("Access-Control-Allow-Origin", "*")
	r.URL.Path = p.ByName("filepath")
	streamFileServerInstance().ServeHTTP(w, r)
}
