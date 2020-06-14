package namenode

import "fmt"

// getDataNodeAddress A function to get a data node address
func (nameNode *NameNode) getNameNodeAddress(dataNode DataNodeData) string {
	return fmt.Sprintf("%s:%s", dataNode.IP, dataNode.Port)
}
