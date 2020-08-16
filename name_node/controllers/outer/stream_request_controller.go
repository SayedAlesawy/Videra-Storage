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

var streamControllerLogPrefix = "[Stream-Controller]"

type clipResultInfo struct {
	Tag       string `json:"tag"`
	StartTime uint64 `json:"start"`
	EndTime   uint64 `json:"end"`
}

// streamResult Represents the result payload of the stream endpoint
type streamResult struct {
	DataNodeID string           `json:"-"`
	VideoLink  string           `json:"src_link"`
	Progress   int              `json:"progress"`
	Clips      []clipResultInfo `json:"clips"`
}

// StreamRequestHandler Handles client's stream request
func (server *Server) StreamRequestHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	log.Println(streamControllerLogPrefix, "Received stream request")

	w.Header().Set("content-type", "application/json")

	expectedParams := []string{"token", "tag"}
	optionalParams := []string{"start", "end"}

	err := requests.ValidateQuery(r.URL.Query(), expectedParams...)
	if errors.IsError(err) {
		log.Println(streamControllerLogPrefix, r.RemoteAddr, err)
		requests.HandleRequestError(w, http.StatusBadRequest, err.Error())

		return
	}

	var result streamResult

	token := r.URL.Query().Get("token")
	tag := r.URL.Query().Get("tag")

	err = requests.ValidateQuery(r.URL.Query(), optionalParams...)
	if !errors.IsError(err) {
		start, startErr := strconv.ParseUint(r.URL.Query().Get("start"), 10, 64)
		end, endErr := strconv.ParseUint(r.URL.Query().Get("end"), 10, 64)

		if errors.IsError(startErr) || errors.IsError(endErr) {
			log.Println(streamControllerLogPrefix, r.RemoteAddr, "Error while parsing start or end times")
			requests.HandleRequestError(w, http.StatusBadRequest, "Error while parsing start or end times")

			return
		}

		if start > end {
			log.Println(streamControllerLogPrefix, r.RemoteAddr, "Malformed request, start can't be > end")
			requests.HandleRequestError(w, http.StatusBadRequest, "Start time can't be greater than end time")

			return
		}
		result.Progress = retrieveIngestionStatus(token)
		result.Clips = retrieveClips(token, tag, start, end)
	} else {
		result.Progress = retrieveIngestionStatus(token)
		result.Clips = retrieveClips(token, tag)
	}

	videoInfo := retrieveVideoInfo(token)
	result.VideoLink = getVideoURL(videoInfo.VideoLink, videoInfo.DataNodeID)

	resp, err := json.Marshal(result)
	if errors.IsError(err) {
		log.Println(streamControllerLogPrefix, r.RemoteAddr, err)
		requests.HandleRequestError(w, http.StatusInternalServerError, err.Error())

		return
	}

	w.Write(resp)
}

// retrieveIngestionStatus retrieves whether ingestion is complete or incomplete
func retrieveIngestionStatus(token string) int {

	queryResult := struct {
		TotalJobCount  int
		TotalDoneCount int
	}{}
	namenode.NodeInstance().DB.Connection.Raw(`
	SELECT total_job_count, total_done_count
	FROM files 
	WHERE files.token = ?`, token).Scan(&queryResult)

	if queryResult.TotalJobCount != 0 {
		return 100 * queryResult.TotalDoneCount / queryResult.TotalJobCount
	}

	return 0
}

// retrieveClips A function to query the clips table for matching records
func retrieveClips(params ...interface{}) []clipResultInfo {
	var clips []clipResultInfo

	if len(params) == 2 {
		namenode.NodeInstance().DB.Connection.Raw("SELECT DISTINCT start_time, end_time FROM clips WHERE token = ? and tag = ? ORDER BY start_time", params[0], params[1]).Scan(&clips)
	} else {
		namenode.NodeInstance().DB.Connection.Raw("SELECT DISTINCT start_time, end_time FROM clips WHERE token = ? and tag = ? and start_time >= ? and start_time <= ? ORDER BY start_time",
			params[0], params[1], params[2], params[3]).Scan(&clips)
	}

	return clips
}

// retrieveVideoInfo A function to query the video info of a file
func retrieveVideoInfo(token string) streamResult {
	videoInfo := streamResult{}
	namenode.NodeInstance().DB.Connection.Raw(`
	SELECT hls_path AS video_link, data_node_id
	FROM files 
	WHERE files.parent = ? and files.token != files.parent`, token).Scan(&videoInfo)

	return videoInfo
}

// getVideoURL updates video url based on datanode url
func getVideoURL(videoPath string, datanodeID string) string {
	if videoPath == "" {
		return ""
	}
	datanodes := namenode.NodeInstance().GetAllDataNodeData()
	for _, datanode := range datanodes {
		if datanode.ID == datanodeID {
			return fmt.Sprintf("%s/%s", namenode.GetURL(datanode.IP, datanode.Port), videoPath)
		}
	}
	return ""
}
