package replication

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/SayedAlesawy/Videra-Storage/config"
)

// for logging hierarchy
var replicationLogPrefix = "[Replication]"

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
		log.Println(replicationLogPrefix, "Undefined request", reqType)
		return errors.New("Undefined request")
	}
}

func replicateInitialRequest(r *http.Request, token string) error {
	config := config.ConfigurationManagerInstance("").DataNodeConfig()
	replicationNode, err := getReplicationNode(token, "init")
	if err != nil {
		log.Println(replicationLogPrefix, err)
		return err
	}
	fmt.Println(replicationLogPrefix, "Sending initial replicate request to ", replicationNode.URL)

	client := newClient(config.ReplicationNumberOfRetries, config.ReplicationWaitingTime)
	req, _ := http.NewRequest(http.MethodPost, replicationNode.URL, nil)
	req.Header = r.Header.Clone()
	req.Header.Set("Parent", token)
	res, err := client.Do(req)
	if err != nil {
		log.Println(replicationLogPrefix, err)
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
	config := config.ConfigurationManagerInstance("").DataNodeConfig()
	replicationNode, err := getReplicationNode(token, "append")
	if err != nil {
		log.Println(replicationLogPrefix, err)
		return err
	}
	client := newClient(config.ReplicationNumberOfRetries, config.ReplicationWaitingTime)
	// for some reason, I can't pass old request body to new request
	// I have to recreate new body
	content, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(replicationLogPrefix, err)
		return err
	}

	body := bytes.NewReader(content)

	req, _ := http.NewRequest(http.MethodPost, replicationNode.URL, body)
	req.Header = r.Header.Clone()
	req.Header.Set("ID", replicationNode.ID)

	res, err := client.Do(req)
	if err != nil {
		cleanUp(token)
		log.Println(replicationLogPrefix, err)
		return err
	}

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		return errors.New("Replication request denied")
	}
	// Replication is finished
	if res.StatusCode == http.StatusCreated {
		cleanUp(token)
	}
	return nil
}

func getReplicationNode(token string, reqType string) (Replica, error) {
	if reqType == "init" {
		replica := Replica{}

		url, err := getAvailableNode()
		if err != nil {
			return replica, err
		}
		replica.URL = url
		replicaJSON, err := encodeReplicaNode(replica)
		if err != nil {
			return replica, err
		}

		key := getReplicaKey()
		field := getReplicaField(token)
		insertIntoHash(key, field, replicaJSON)
		log.Println(replicationLogPrefix, fmt.Sprintf("Replication node for file %s is set to URL: %s", token, url))
		return replica, nil
	} else if reqType == "append" {
		key := getReplicaKey()
		field := getReplicaField(token)
		node, err := getFromHash(key, field)

		// an error has happend
		if invalidCacheValue(node, err) {
			log.Println(replicationLogPrefix, fmt.Sprintf("Cache missed for file %s", token))
			return Replica{}, errors.New("Cache missed for file")
		}

		replica, err := decodeReplicaNodeData(node)
		if err != nil {
			return Replica{}, err
		}

		return replica, nil
	} else {
		return Replica{}, errors.New("Invalid request type")
	}
}

func getAvailableNode() (string, error) {
	dataNodeConfig := config.ConfigurationManagerInstance("").DataNodeConfig()
	client := newClient(dataNodeConfig.ReplicationNumberOfRetries, dataNodeConfig.ReplicationWaitingTime)
	nodeID := bytes.NewReader(([]byte(dataNodeConfig.ID)))
	req, _ := http.NewRequest(http.MethodGet, dataNodeConfig.NameNodeReplicationURL, nodeID)
	res, err := client.Do(req)
	if err != nil {
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
	key := getReplicaKey()
	field := getReplicaField(token)
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
