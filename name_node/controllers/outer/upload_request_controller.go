package outer

import (
	"fmt"
	"log"
	"net/http"

	namenode "github.com/SayedAlesawy/Videra-Storage/name_node"
	"github.com/SayedAlesawy/Videra-Storage/utils/errors"
	"github.com/julienschmidt/httprouter"
)

// ServeUploadURL Handles client's request to retrieve data node url for uploading
func (server *Server) ServeUploadURL(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// TODO: Add check if current name node is not the master and return the master ip

	nameNode := namenode.NodeInstance()

	// Get node with minimum number of clients requests
	choosenDataNode, err := server.getAvailableDataNode(nameNode)

	// There's no available nodes
	if errors.IsError(err) {
		log.Println(logPrefix, r.RemoteAddr, err)
		w.Write([]byte("Service Unavailable"))
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	// update request count for further selection
	choosenDataNode.RequestCount++
	nameNode.InsertDataNodeData(choosenDataNode)

	choosenNodeURL := server.getDataNodeUploadURL(choosenDataNode.IP, choosenDataNode.ExternalPort)
	log.Println(logPrefix, r.RemoteAddr, fmt.Sprintf("routed to node %s with endpoint %s", choosenDataNode.ID, choosenNodeURL))
	w.Write([]byte(choosenNodeURL))
	w.WriteHeader(http.StatusOK)
}

// getAvailableDataNode is a function to get available data node
// it tries to choose a data node with minimum load
func (server *Server) getAvailableDataNode(nameNode *namenode.NameNode) (namenode.DataNodeData, error) {
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
		return choosenDataNode, errors.New("No datanodes available")
	}

	return choosenDataNode, nil
}

// getAddress A function to get the address on which the internal controller listens
func (server *Server) getDataNodeUploadURL(ip string, port string) string {
	return fmt.Sprintf("http://%s:%s/upload", ip, port)
}
