package outer

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/SayedAlesawy/Videra-Storage/config"
	"github.com/julienschmidt/httprouter"
)

// Server Handles external communication routers with client
type Server struct {
	IP   string //IP on which the server is hosted
	Port string //Port on which the server listens to requests
}

// logPrefix Used for hierarchical logging
var logPrefix = "[External-Controller]"

// serverOnce Used to garauntee thread safety for singleton instances
var serverOnce sync.Once

// serverInstance A singleton instance of the server object
var serverInstance *Server

// ServerInstance A function to return a singleton server instance
func ServerInstance() *Server {
	dataNodeConfig := config.ConfigurationManagerInstance("").DataNodeConfig()

	serverOnce.Do(func() {
		server := Server{
			IP:   dataNodeConfig.IP,
			Port: dataNodeConfig.Port,
		}

		serverInstance = &server
	})

	return serverInstance
}

// Start A function to start the external controllers server
func (server *Server) Start() {
	router := httprouter.New()
	router.POST("/upload", server.UploadRequestHandler)
	router.GET("/stream/*filepath", server.StreamingHandler)
	router.GET("/thumbnail/*filepath", server.ThumbnailsHandler)
	address := server.getAddress()

	log.Println(logPrefix, fmt.Sprintf("Listening for external requests on %s", address))
	log.Fatal(http.ListenAndServe(address, router))
}

// getAddress A function to get the address on which the external controller listens
func (server *Server) getAddress() string {
	return fmt.Sprintf("%s:%s", server.IP, server.Port)
}
