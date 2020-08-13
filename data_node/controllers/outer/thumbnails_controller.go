package outer

import (
	"net/http"
	"sync"

	"github.com/SayedAlesawy/Videra-Storage/config"
	"github.com/julienschmidt/httprouter"
)

// thumbnailFileServerOnce Used to garauntee thread safety for singleton instances
var thumbnailFileServerOnce sync.Once

// thumbnailFileServerInstance A singleton instance of the thumbnailFileServer object, to serve files in directory
var thumbnailFileServer http.Handler

// thumbnailFileServerInstance A function to return a singleton thumbnailFileServer instance
func thumbnailFileServerInstance() http.Handler {
	dataNodeConfig := config.ConfigurationManagerInstance("").DataNodeConfig()

	thumbnailFileServerOnce.Do(func() {
		thumbnailFileServer = http.FileServer(http.Dir(dataNodeConfig.ThumbnailFolderName))
	})

	return thumbnailFileServer
}

// ThumbnailsHandler is a handle responsible for serving thumbnails retrieval requests
func (server *Server) ThumbnailsHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// Some browsers doesn't allow reading files from other directory
	// so we can set this header, or allow CORS in the browser
	// but this header isn't recommended in production
	w.Header().Set("Access-Control-Allow-Origin", "*")
	r.URL.Path = p.ByName("filepath")
	thumbnailFileServerInstance().ServeHTTP(w, r)
}
