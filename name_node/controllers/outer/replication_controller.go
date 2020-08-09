package outer

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	namenode "github.com/SayedAlesawy/Videra-Storage/name_node"
	"github.com/SayedAlesawy/Videra-Storage/utils/errors"
	"github.com/julienschmidt/httprouter"
)

// ReplicationAddressesHandler is a handler responsible for providing data node addresses for replication
func (server *Server) ReplicationAddressesHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	// Get available node for replication
	chosenDataNode, err := server.getReplicationNode(r.RemoteAddr)
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
func (server *Server) getReplicationNode(originalHost string) (namenode.DataNodeData, error) {
	nameNode := namenode.NodeInstance()
	var chosenDataNode namenode.DataNodeData

	dataNodesData := nameNode.GetAllDataNodeData()

	// There's no available nodes
	if len(dataNodesData) <= 1 {
		return chosenDataNode, errors.New("No datanodes available")
	}

	hostIdx := getNodeByIP(dataNodesData, getNodeIP(originalHost))

	// for some reason, the requester node is not available
	if hostIdx == -1 {
		return chosenDataNode, errors.New("Invalid IP")
	}

	// Request is routed to the node next to data node
	chosenDataNode = dataNodesData[(hostIdx+1)%len(dataNodesData)]
	return chosenDataNode, nil
}

// getNodeByIP gets index of node with ip
func getNodeByIP(dataNodes []namenode.DataNodeData, ip string) int {
	log.Println("Node to get", ip)
	for idx, datanode := range dataNodes {
		if datanode.IP == ip {
			return idx
		}
	}

	return -1
}

// getNodeIP removes port part from ip, for example 192.12.1.1:8080 to 192.12.1.1
func getNodeIP(URL string) string {
	idx := strings.Index(URL, ":")
	if idx != -1 {
		URL = URL[:idx]
	}
	return URL
}
