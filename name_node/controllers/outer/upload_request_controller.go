package outer

import (
	"fmt"
	"log"
	"net/http"

	namenode "github.com/SayedAlesawy/Videra-Storage/name_node"
	"github.com/SayedAlesawy/Videra-Storage/utils/errors"
	"github.com/julienschmidt/httprouter"
)

// UploadRequestHandler Handles client's request to retrieve data node url for uploading
func (server *Server) UploadRequestHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	nameNode := namenode.NodeInstance()

	// Get node with minimum number of clients requests
	chosenDataNode, err := server.getAvailableDataNode(nameNode)
	// There's no available nodes
	if errors.IsError(err) {
		log.Println(logPrefix, r.RemoteAddr, err)
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("Service Unavailable"))
		return
	}

	// update request count for further selection
	chosenDataNode.RequestCount++
	nameNode.InsertDataNodeData(chosenDataNode)

	chosenNodeURL := server.getDataNodeUploadURL(chosenDataNode.IP, chosenDataNode.Port)
	log.Println(logPrefix, r.RemoteAddr, fmt.Sprintf("routed to node %s with endpoint %s", chosenDataNode.ID, chosenNodeURL))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(chosenNodeURL))
}

// getAvailableDataNode is a function to get available data node
// it tries to choose a data node with minimum load
func (server *Server) getAvailableDataNode(nameNode *namenode.NameNode) (namenode.DataNodeData, error) {
	var chosenDataNode namenode.DataNodeData

	dataNodesData := nameNode.GetAllDataNodeData()
	// There's no available nodes
	if len(dataNodesData) == 0 {
		return chosenDataNode, errors.New("No datanodes available")
	}

	for idx, dataNode := range dataNodesData {
		if idx == 0 {
			chosenDataNode = dataNode
		} else {
			if chosenDataNode.RequestCount > dataNode.RequestCount {
				chosenDataNode = dataNode
			}
		}
	}
	return chosenDataNode, nil
}

// getAddress A function to get the address on which the internal controller listens
func (server *Server) getDataNodeUploadURL(ip string, port string) string {
	return fmt.Sprintf("http://%s:%s/upload", ip, port)
}
