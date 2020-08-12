package outer

import (
	"net/http"
	"sync"

	"github.com/SayedAlesawy/Videra-Storage/config"
	"github.com/julienschmidt/httprouter"
)

// fileServerOnce Used to garauntee thread safety for singleton instances
var fileServerOnce sync.Once

// fileServerInstance A singleton instance of the fileServer object, to serve files in directory
var fileServer http.Handler

// fileServerInstance A function to return a singleton fileServer instance
func fileServerInstance() http.Handler {
	dataNodeConfig := config.ConfigurationManagerInstance("").DataNodeConfig()

	fileServerOnce.Do(func() {
		fileServer = http.FileServer(http.Dir(dataNodeConfig.StreamFolderName))
	})

	return fileServer
}

// StreamingHandler is a handle responsible for serving streaming requests
func (server *Server) StreamingHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// Some browsers doesn't allow reading files from other directory
	// so we can set this header, or allow CORS in the browser
	// but this header isn't recommended in production
	w.Header().Set("Access-Control-Allow-Origin", "*")
	r.URL.Path = p.ByName("filepath")
	fileServerInstance().ServeHTTP(w, r)
}
