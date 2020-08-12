package replication

import (
	"fmt"

	datanode "github.com/SayedAlesawy/Videra-Storage/data_node"
)

// insertIntoHash A function to insert a new entry in a redis hash
func insertIntoHash(key string, field string, value string) error {
	return datanode.NodeInstance().Cache.HSet(key, field, value).Err()
}

// deleteFromHash A function to delete a certain field in redis hash
func deleteFromHash(key string, field string) error {
	return datanode.NodeInstance().Cache.HDel(key, field).Err()
}

// getFromHash A function to field from a redis hash
func getFromHash(key string, field string) (string, error) {
	return datanode.NodeInstance().Cache.HGet(key, field).Result()
}

// invalidCacheValue checks the value returned from cache is invalid
func invalidCacheValue(value string, err error) bool {
	return fmt.Sprintf("%v", err) == "redis: nil" && value == ""
}

// getReplicaKey is a helper function to generate replica key for cache
func getReplicaKey() string {
	nodeID := datanode.NodeInstance().ID
	return fmt.Sprintf("DN-%s-replicas", nodeID)

}

// getReplicaKey is a helper function to generate replica field for cache
func getReplicaField(token string) string {
	return fmt.Sprintf("replica-%s", token)
}
