package replication

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/SayedAlesawy/Videra-Storage/config"
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

// TODO remove this
var rr = Replica{}

// newClient is a function that returns customized http client
func newClient(maxRetries int, waitingTime int) *http.Client {
	clientretry := retryablehttp.NewClient()
	clientretry.RetryMax = maxRetries
	clientretry.RetryWaitMin = time.Duration(time.Duration(waitingTime) * time.Second)
	clientretry.RetryWaitMax = time.Duration(time.Duration(waitingTime) * time.Second)

	return clientretry.StandardClient()
}

// ReplicateVideo is responsible for replicating video to other data nodes
func ReplicateVideo(r *http.Request, token string) {
	reqType := strings.ToLower(r.Header.Get("Request-Type"))

	switch reqType {
	case "init":
		replicateInitialRequest(r, token)
	case "append":
		replicateAppendRequest(r, token)
	default:
		log.Println("Undefined req")
	}
}

func replicateInitialRequest(r *http.Request, token string) {
	replicationNode := getReplicationNode(token)
	fmt.Println(replicationLogPrefix, "Sending initial replicate request to ", replicationNode.URL)
	client := newClient(retries, waitingTime)
	req, _ := http.NewRequest(http.MethodPost, replicationNode.URL, nil)
	req.Header = r.Header.Clone()
	req.Header.Set("Parent", token)
	res, _ := client.Do(req)
	fileID := res.Header.Get("ID")
	updateReplicaID(token, fileID)
}

func replicateAppendRequest(r *http.Request, token string) {
	replicationNode := getReplicationNode(token)
	client := newClient(retries, waitingTime)
	// for some reason, I can't pass old request body to new request
	// I have to recreate new body
	content, _ := ioutil.ReadAll(r.Body)
	body := bytes.NewReader(content)

	req, _ := http.NewRequest(http.MethodPost, replicationNode.URL, body)
	req.Header = r.Header.Clone()
	req.Header.Set("ID", replicationNode.ID)

	_, _ = client.Do(req)
}

func getReplicationNode(token string) Replica {
	// key := getReplicaKey(token)
	// retrieve replica from redis using key
	// if not exist, request from name node and set in redis

	// TODO: Remove this
	if rr.URL == "" {
		url, _ := getAvailableNode()
		rr.URL = url
	}
	////////////////////////
	return rr
}

func getAvailableNode() (string, error) {
	dataNodeConfig := config.ConfigurationManagerInstance("").DataNodeConfig()
	client := newClient(retries, waitingTime)
	req, _ := http.NewRequest(http.MethodGet, dataNodeConfig.NameNodeReplicationURL, nil)
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

func getReplicaKey(token string) string {
	return fmt.Sprintf("replica-%s", token)
}

func updateReplicaID(token string, fileID string) {

	// key := getReplicaKey(token)
	// retrieve replica from redis using key
	// Update ID in replica to fileID
	rr.ID = fileID
}
