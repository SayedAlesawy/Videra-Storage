package outer

import (
	"fmt"
	"net/http"

	namenode "github.com/SayedAlesawy/Videra-Storage/name_node"
	"github.com/julienschmidt/httprouter"
)

// ServeUploadURL Handles client's request to retrieve data node url for uploading
func (server *Server) ServeUploadURL(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// TODO: Add check if current name node is not the master and return the master ip

	nameNode := namenode.NodeInstance()

	// Get node with minimum number of clients requests
	var choosenDataNode namenode.DataNodeData
	for idx, dataNode := range nameNode.GetAllDataNodeData() {
		if idx == 0 {
			choosenDataNode = dataNode
		} else {
			if choosenDataNode.RequestCount > dataNode.RequestCount {
				choosenDataNode = dataNode
			}
		}
	}
	// There's no available nodes
	if choosenDataNode.IP == "" {
		w.Write([]byte("Service Unavailable"))
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	choosenDataNode.RequestCount++
	nameNode.InsertDataNodeData(choosenDataNode)

	w.Write([]byte(server.getDataNodeUploadURL(choosenDataNode.IP, choosenDataNode.ExternalPort)))
	w.WriteHeader(http.StatusOK)
}

// getAddress A function to get the address on which the internal controller listens
func (server *Server) getDataNodeUploadURL(ip string, port string) string {
	return fmt.Sprintf("http://%s:%s/upload", ip, port)
}
