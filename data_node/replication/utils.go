package replication

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

// newClient is a function that returns customized http client
func newClient(maxRetries int, waitingTime int) *http.Client {
	clientretry := retryablehttp.NewClient()
	clientretry.RetryMax = maxRetries
	clientretry.RetryWaitMin = time.Duration(time.Duration(waitingTime) * time.Second)
	clientretry.RetryWaitMax = time.Duration(time.Duration(waitingTime) * time.Second)

	return clientretry.StandardClient()
}

// encode A function to encode the data node data into json format
func encodeReplicaNode(replica Replica) (string, error) {
	encodedData, err := json.Marshal(replica)
	if err != nil {
		return "", err
	}

	return string(encodedData), nil
}

// decodeReplicaNodeData Decodes the stringified data node data
func decodeReplicaNodeData(encodedData string) (Replica, error) {
	var dataNodeData Replica

	err := json.Unmarshal([]byte(encodedData), &dataNodeData)
	if err != nil {
		return Replica{}, err
	}

	return dataNodeData, nil
}

func cleanUp(id string) {
	key := getReplicaKey()
	field := getReplicaField(id)
	deleteFromHash(key, field)
}
