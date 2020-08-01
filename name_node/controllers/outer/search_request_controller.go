package outer

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	namenode "github.com/SayedAlesawy/Videra-Storage/name_node"
	"github.com/SayedAlesawy/Videra-Storage/utils/errors"
	"github.com/SayedAlesawy/Videra-Storage/utils/requests"
	"github.com/julienschmidt/httprouter"
)

var scLogPrefix = "[Search-Controller]"

// SearchRequestHandler Handles client's search request
func (server *Server) SearchRequestHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	log.Println(scLogPrefix, "Received search request")

	w.Header().Set("content-type", "application/json")

	expectedHeaders := []string{"Tag"}
	optionalHeaders := []string{"StartTime", "EndTime"}

	err := requests.ValidateUploadHeaders(&r.Header, expectedHeaders...)
	if errors.IsError(err) {
		log.Println(scLogPrefix, r.RemoteAddr, err)
		requests.HandleRequestError(w, http.StatusBadRequest, err.Error())

		return
	}

	var clips []namenode.Clip

	tag := r.Header.Get("Tag")

	err = requests.ValidateUploadHeaders(&r.Header, optionalHeaders...)
	if !errors.IsError(err) {
		start, startErr := strconv.ParseUint(r.Header.Get("StartTime"), 10, 64)
		end, endErr := strconv.ParseUint(r.Header.Get("EndTime"), 10, 64)

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

		clips = retrieveClips(tag, start, end)
	} else {
		clips = retrieveClips(tag)
	}

	resp, err := json.Marshal(clips)
	if errors.IsError(err) {
		log.Println(scLogPrefix, r.RemoteAddr, err)
		requests.HandleRequestError(w, http.StatusInternalServerError, err.Error())

		return
	}

	w.Write(resp)
}

// retrieveClips A function to query the clips table for matching records
func retrieveClips(params ...interface{}) []namenode.Clip {
	var clips []namenode.Clip

	if len(params) == 1 {
		namenode.NodeInstance().DB.Connection.Where("tag = ?", params[0]).Find(&clips)
	} else {
		namenode.NodeInstance().DB.Connection.Where("tag = ? and start_time >= ? and end_time <= ?",
			params[0], params[1], params[2]).Find(&clips)
	}

	return clips
}
