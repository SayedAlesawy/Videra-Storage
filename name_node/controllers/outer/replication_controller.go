package outer

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	namenode "github.com/SayedAlesawy/Videra-Storage/name_node"
	"github.com/SayedAlesawy/Videra-Storage/utils/errors"
	"github.com/julienschmidt/httprouter"
)

// ReplicationAddressesHandler is a handler responsible for providing data node addresses for replication
func (server *Server) ReplicationAddressesHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	body, err := ioutil.ReadAll(r.Body)
	if errors.IsError(err) {
		log.Println(logPrefix, r.RemoteAddr, err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
		return
	}

	// Get available node for replication
	chosenDataNode, err := server.getReplicationNode(string(body))
	// There's no available nodes
	if errors.IsError(err) {
		log.Println(logPrefix, r.RemoteAddr, err)
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("Service Unavailable"))
		return
	}

	chosenNodeURL := server.getDataNodeUploadURL(chosenDataNode.IP, chosenDataNode.Port)
	log.Println(logPrefix, r.RemoteAddr, fmt.Sprintf("routed replication to node %s with endpoint %s", chosenDataNode.ID, chosenNodeURL))
	w.Write([]byte(chosenNodeURL))
}

// getReplicationNode is a function to get available data node for replication
func (server *Server) getReplicationNode(nodeID string) (namenode.DataNodeData, error) {
	nameNode := namenode.NodeInstance()
	var chosenDataNode namenode.DataNodeData

	dataNodesData := nameNode.GetAllDataNodeData()

	// There's no available nodes
	if len(dataNodesData) <= 1 {
		return chosenDataNode, errors.New("No datanodes available")
	}

	hostIdx := getNodeByID(dataNodesData, nodeID)

	// for some reason, the requester node is not available
	if hostIdx == -1 {
		return chosenDataNode, errors.New("Invalid Node ID")
	}

	// Request is routed to the node next to data node
	chosenDataNode = dataNodesData[(hostIdx+1)%len(dataNodesData)]
	return chosenDataNode, nil
}

// getNodeByID gets index of node with id
func getNodeByID(dataNodes []namenode.DataNodeData, id string) int {
	for idx, datanode := range dataNodes {
		if datanode.ID == id {
			return idx
		}
	}

	return -1
}
