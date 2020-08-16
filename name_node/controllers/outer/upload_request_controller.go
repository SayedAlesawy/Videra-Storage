package outer

import (
	"fmt"
	"log"
	"net/http"
	"time"

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
	chosenDataNode.LastRequestTime = time.Now()
	nameNode.InsertDataNodeData(chosenDataNode)

	chosenNodeURL := server.getDataNodeUploadURL(chosenDataNode.IP, chosenDataNode.Port)
	log.Println(logPrefix, r.RemoteAddr, fmt.Sprintf("routed to node %s with endpoint %s", chosenDataNode.ID, chosenNodeURL))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(chosenNodeURL))
}

// getAvailableDataNode is a function to get available data node
// it tries to choose a data node with minimum load
func (server *Server) getAvailableDataNode(nameNode *namenode.NameNode) (namenode.DataNodeData, error) {
	dataNodes := nameNode.GetAllDataNodeData()
	// There's no available nodes
	if len(dataNodes) == 0 {
		return namenode.DataNodeData{}, errors.New("No datanodes available")
	}

	var chosenDataNode namenode.DataNodeData
	hasChoosenNode := false

	for _, dataNode := range dataNodes {
		if !dataNode.GPU {
			continue
		}

		if hasChoosenNode {
			if dataNode.LastRequestTime.Before(chosenDataNode.LastRequestTime) {
				chosenDataNode = dataNode
			}
		} else {
			chosenDataNode = dataNode
			hasChoosenNode = true
		}
	}

	if !hasChoosenNode {
		return namenode.DataNodeData{}, errors.New("Can't find a machine with GPU")
	}

	return chosenDataNode, nil
}

// getAddress A function to get the address on which the internal controller listens
func (server *Server) getDataNodeUploadURL(ip string, port string) string {
	return fmt.Sprintf("http://%s:%s/upload", ip, port)
}
