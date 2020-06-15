package namenode

import (
	"fmt"
	"log"

	"github.com/SayedAlesawy/Videra-Storage/utils/errors"
)

// InsertDataNodeData A function to insert data node data into active nodes hash
func (nameNode *NameNode) InsertDataNodeData(dataNodeData DataNodeData) bool {
	encodedData, err := dataNodeData.encode()
	if errors.IsError(err) {
		log.Println(logPrefix, "Unable to marshal data node data", dataNodeData)

		return false
	}

	err = nameNode.insertIntoHash(nameNode.dataNodesTrackingKey, dataNodeData.ID, encodedData)
	if errors.IsError(err) {
		log.Println(logPrefix, fmt.Sprintf("Unable to insert into redis hash: %s for data node: %s",
			nameNode.dataNodesTrackingKey, dataNodeData.ID))

		return false
	}

	return true
}

// RemoveDataNodeData A function to remove data node data from active nodes hash
func (nameNode *NameNode) RemoveDataNodeData(dataNodeData DataNodeData) {
	err := nameNode.deleteFromHash(nameNode.dataNodesTrackingKey, dataNodeData.ID)
	if errors.IsError(err) {
		log.Println(logPrefix, fmt.Sprintf("Unable to remove from redis hash: %s for data node: %s",
			nameNode.dataNodesTrackingKey, dataNodeData.ID))
	}
}

// GetAllDataNodeData A function to get all data node data from active hash
func (nameNode *NameNode) GetAllDataNodeData() []DataNodeData {
	var dataNodes []DataNodeData

	nodes, err := nameNode.getAllFromHash(nameNode.dataNodesTrackingKey)
	if errors.IsError(err) {
		log.Println(logPrefix, "Unable to fetch data node info from redis")

		return dataNodes
	}

	for _, node := range nodes {
		decodedNode, err := nameNode.decodeDataNodeData(node)
		if errors.IsError(err) {
			log.Println(logPrefix, "Unable to decode data node data", node)

			continue
		}

		dataNodes = append(dataNodes, decodedNode)
	}

	return dataNodes
}

// insertIntoHash A function to insert a new entry in a redis hash
func (nameNode *NameNode) insertIntoHash(key string, field string, value string) error {
	return nameNode.cache.HSet(key, field, value).Err()
}

// deleteFromHash A function to delete a certain field in redis hash
func (nameNode *NameNode) deleteFromHash(key string, field ...string) error {
	return nameNode.cache.HDel(key, field...).Err()
}

// getAllFromHash A function to get all fields from a redis hash
func (nameNode *NameNode) getAllFromHash(key string) (map[string]string, error) {
	return nameNode.cache.HGetAll(key).Result()
}
