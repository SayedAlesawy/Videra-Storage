package replication

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/SayedAlesawy/Videra-Storage/config"
	datanode "github.com/SayedAlesawy/Videra-Storage/data_node"
	"github.com/hashicorp/go-retryablehttp"
)

// for logging hierarchy
var replicationLogPrefix = "[Replication]"

// number if retries in case of failure
var retries = 3

// waiting time between retries
var waitingTime = 3

// Replica is a model for replication mode info
type Replica struct {
	URL string //URL to send data to
	ID  string //File ID to append to
}

// newClient is a function that returns customized http client
func newClient(maxRetries int, waitingTime int) *http.Client {
	clientretry := retryablehttp.NewClient()
	clientretry.RetryMax = maxRetries
	clientretry.RetryWaitMin = time.Duration(time.Duration(waitingTime) * time.Second)
	clientretry.RetryWaitMax = time.Duration(time.Duration(waitingTime) * time.Second)

	return clientretry.StandardClient()
}

// ReplicateVideo is responsible for replicating video to other data nodes
func ReplicateVideo(r *http.Request, token string) error {
	reqType := strings.ToLower(r.Header.Get("Request-Type"))

	switch reqType {
	case "init":
		err := replicateInitialRequest(r, token)
		return err
	case "append":
		err := replicateAppendRequest(r, token)
		return err
	default:
		log.Println("Undefined req")
		return errors.New("Undefined request")
	}
}

func replicateInitialRequest(r *http.Request, token string) error {
	replicationNode, err := getReplicationNode(token)
	if err != nil {
		return err
	}
	fmt.Println(replicationLogPrefix, "Sending initial replicate request to ", replicationNode.URL)

	client := newClient(retries, waitingTime)
	req, _ := http.NewRequest(http.MethodPost, replicationNode.URL, nil)
	req.Header = r.Header.Clone()
	req.Header.Set("Parent", token)
	res, err := client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusCreated {
		return errors.New("Initial replicate request denied")
	}
	fileID := res.Header.Get("ID")
	updateReplicaID(replicationNode, token, fileID)
	return nil
}

func replicateAppendRequest(r *http.Request, token string) error {
	replicationNode, err := getReplicationNode(token)
	if err != nil {
		return err
	}
	client := newClient(retries, waitingTime)
	// for some reason, I can't pass old request body to new request
	// I have to recreate new body
	content, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	body := bytes.NewReader(content)

	req, _ := http.NewRequest(http.MethodPost, replicationNode.URL, body)
	req.Header = r.Header.Clone()
	req.Header.Set("ID", replicationNode.ID)

	res, err := client.Do(req)
	if err != nil {
		deleteWithHash(getReplicaKey(token))
		return err
	}

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		return errors.New("Replication request denied")
	}
	// Replication is finished
	if res.StatusCode == http.StatusCreated {
		deleteWithHash(getReplicaKey(token))
	}
	return nil
}

func getReplicationNode(token string) (Replica, error) {
	key := getReplicaKey(token)
	node, _ := getFromHash(key, token)

	replica := Replica{}

	// not exist in cache
	if node == "" {
		url, err := getAvailableNode()
		if err != nil {
			return replica, err
		}
		replica.URL = url
		replicaJSON, err := encodeReplicaNode(replica)
		if err != nil {
			return replica, err
		}
		insertIntoHash(key, token, replicaJSON)
		log.Println(replicationLogPrefix, "Replica not exist in cache, now is set to URL: ", url)
	} else {
		decodedReplica, err := decodeReplicaNodeData(node)
		if err != nil {
			return replica, err
		}
		replica = decodedReplica
	}

	return replica, nil
}

func getAvailableNode() (string, error) {
	dataNodeConfig := config.ConfigurationManagerInstance("").DataNodeConfig()
	client := newClient(retries, waitingTime)
	nodeID := bytes.NewReader(([]byte(dataNodeConfig.ID)))
	req, _ := http.NewRequest(http.MethodGet, dataNodeConfig.NameNodeReplicationURL, nodeID)
	res, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return "", err
	}

	defer res.Body.Close()

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		return "", err
	}

	body := string(bodyBytes)

	if res.StatusCode != http.StatusOK {
		log.Println(body)
		return "", errors.New(body)
	}

	return body, nil
}

func updateReplicaID(replica Replica, token string, fileID string) error {
	key := getReplicaKey(token)
	field := token
	replica.ID = fileID
	replicaJSON, err := encodeReplicaNode(replica)
	if err != nil {
		return err
	}

	err = insertIntoHash(key, field, replicaJSON)
	if err != nil {
		return err
	}

	return nil
}

// insertIntoHash A function to insert a new entry in a redis hash
func insertIntoHash(key string, field string, value string) error {
	return datanode.NodeInstance().Cache.HSet(key, field, value).Err()
}

// deleteWithHash A function to delete a certain field in redis hash
func deleteWithHash(key string) error {
	log.Println(key)
	return datanode.NodeInstance().Cache.Del(key).Err()
}

// getFromHash A function to field from a redis hash
func getFromHash(key string, field string) (string, error) {
	return datanode.NodeInstance().Cache.HGet(key, field).Result()
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

func getReplicaKey(token string) string {
	return fmt.Sprintf("replica-%s", token)
}
