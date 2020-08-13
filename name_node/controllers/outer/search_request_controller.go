package outer

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	namenode "github.com/SayedAlesawy/Videra-Storage/name_node"
	"github.com/SayedAlesawy/Videra-Storage/utils/errors"
	"github.com/SayedAlesawy/Videra-Storage/utils/requests"
	"github.com/julienschmidt/httprouter"
)

var scLogPrefix = "[Search-Controller]"

// searchResult Represents the result payload of the search endpoint
type searchResult struct {
	DataNodeID    string `json:"-"`
	Name          string `json:"name"`
	Token         string `json:"token"`
	ThumbnailPath string `json:"thumbnail"`
}

// SearchRequestHandler Handles client's search request
func (server *Server) SearchRequestHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	log.Println(scLogPrefix, "Received search request")

	w.Header().Set("content-type", "application/json")

	expectedParams := []string{"tag"}
	optionalParams := []string{"start", "end"}

	err := requests.ValidateQuery(r.URL.Query(), expectedParams...)
	if errors.IsError(err) {
		log.Println(scLogPrefix, r.RemoteAddr, err)
		requests.HandleRequestError(w, http.StatusBadRequest, err.Error())

		return
	}

	var results []searchResult

	tag := r.URL.Query().Get("tag")

	err = requests.ValidateQuery(r.URL.Query(), optionalParams...)
	if !errors.IsError(err) {
		start, startErr := strconv.ParseUint(r.URL.Query().Get("start"), 10, 64)
		end, endErr := strconv.ParseUint(r.URL.Query().Get("end"), 10, 64)

		if errors.IsError(startErr) || errors.IsError(endErr) {
			log.Println(scLogPrefix, r.RemoteAddr, "Error while parsing start or end times")
			requests.HandleRequestError(w, http.StatusBadRequest, "Error while parsing start or end times")

			return
		}

		if start > end {
			log.Println(scLogPrefix, r.RemoteAddr, "Malformed request, start can't be > end")
			requests.HandleRequestError(w, http.StatusBadRequest, "Start time can't be greater than end time")

			return
		}

		results = retrieveVideos(tag, start, end)
	} else {
		results = retrieveVideos(tag)
	}

	updateThumbnailURL(results)
	resp, err := json.Marshal(results)
	if errors.IsError(err) {
		log.Println(scLogPrefix, r.RemoteAddr, err)
		requests.HandleRequestError(w, http.StatusInternalServerError, err.Error())

		return
	}

	w.Write(resp)
}

// retrieveVideos A function to query the clips table for matching records
func retrieveVideos(params ...interface{}) []searchResult {
	var results []searchResult
	if len(params) == 1 {
		namenode.NodeInstance().DB.Connection.Raw(`
		SELECT files.parent AS token, files.name, files.thumbnail_path, files.ID, files.data_node_id
		FROM files INNER JOIN 
		(
			SELECT DISTINCT(token) FROM clips WHERE tag = ?
		) AS videos 
		ON files.parent =  videos.token 
		WHERE files.parent != files.token`,
			params[0]).Scan(&results)
	} else {
		namenode.NodeInstance().DB.Connection.Raw(`
		SELECT files.parent AS token, files.name, files.thumbnail_path, files.data_node_id AS DataNodeID
		FROM files INNER JOIN 
		(
			SELECT DISTINCT(token) FROM clips WHERE tag = ? and start_time >= ? and start_time <= ?
		) AS videos 
		ON files.parent =  videos.token 
		WHERE files.parent != files.token`,
			params[0], params[1], params[2]).Scan(&results)
	}

	return results
}

// updateThumbnailURL updates thumbnail url based on datanode url
func updateThumbnailURL(results []searchResult) {
	datanodes := namenode.NodeInstance().GetAllDataNodeData()
	URLS := make(map[string]string)
	for _, datanode := range datanodes {
		URLS[datanode.ID] = namenode.GetURL(datanode.IP, datanode.Port)
	}

	for i := range results {
		datanodeURL, ok := URLS[results[i].DataNodeID]
		if ok {
			results[i].ThumbnailPath = fmt.Sprintf("%s/%s", datanodeURL, results[i].ThumbnailPath)
		} else {
			results[i].ThumbnailPath = ""
		}
	}
}
