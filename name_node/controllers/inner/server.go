package inner

import (
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/SayedAlesawy/Videra-Ingestion/orchestrator/utils/errors"
	"github.com/SayedAlesawy/Videra-Storage/protobuf"
	grpc "google.golang.org/grpc"
)

// Server Handles internal comm routers with data nodes
type Server struct {
	IP      string //IP on which the server is hosted
	Port    string //Port on which the server listens to requests
	Network string //Network protocol used by the server
}

// logPrefix Used for hierarchical logging
var logPrefix = "[Internal-Controller]"

// serverOnce Used to garauntee thread safety for singleton instances
var serverOnce sync.Once

// monitorInstance A singleton instance of the server object
var serverInstance *Server

// ServerInstance A function to return a singleton server instance
func ServerInstance() *Server {
	serverOnce.Do(func() {
		server := Server{
			Network: "tcp",
			IP:      "127.0.0.1",
			Port:    "5555",
		}

		serverInstance = &server
	})

	return serverInstance
}

// Start A function to start the internal controllers server
func (server *Server) Start() {
	//Obtain net listener
	listener, err := net.Listen(server.Network, server.getAddress())
	errors.HandleError(err, fmt.Sprintf("%s Unable to start internal controller", logPrefix), true)

	//Start gRPC server
	grpcServer := grpc.NewServer()
	protobuf.RegisterInternalRoutesServer(grpcServer, server)

	//Server gRPC routes on the obtained listener
	log.Println(logPrefix, fmt.Sprintf("Listening on %s", server.getAddress()))

	grpcServer.Serve(listener)
}

func (server *Server) getAddress() string {
	return fmt.Sprintf("%s:%s", server.IP, server.Port)
}
