package namenode

import (
	"encoding/json"
	"fmt"

	"github.com/SayedAlesawy/Videra-Storage/utils/errors"
)

// getDataNodeAddress A function to get a data node address
func (nameNode *NameNode) getNameNodeAddress(dataNode DataNodeData) string {
	return fmt.Sprintf("%s:%s", dataNode.IP, dataNode.Port)
}

// NewDataNodeData A function to obtain a new data node data object
func NewDataNodeData(id string, ip string, port string) DataNodeData {
	return DataNodeData{
		ID:      id,
		IP:      ip,
		Port:    port,
		Latency: 0,
	}
}

// encode A function to encode the data node data into json format
func (dataNodeData *DataNodeData) encode() (string, error) {
	encodedData, err := json.Marshal(dataNodeData)
	if errors.IsError(err) {
		return "", err
	}

	return string(encodedData), nil
}

// decodeDataNodeData Decodes the stringified data node data
func decodeDataNodeData(encodedData string) (DataNodeData, error) {
	var dataNodeData DataNodeData

	err := json.Unmarshal([]byte(encodedData), &dataNodeData)
	if errors.IsError(err) {
		return DataNodeData{}, err
	}

	return dataNodeData, nil
}
