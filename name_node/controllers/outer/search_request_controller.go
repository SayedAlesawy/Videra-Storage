package outer

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	namenode "github.com/SayedAlesawy/Videra-Storage/name_node"
	"github.com/SayedAlesawy/Videra-Storage/utils/errors"
	"github.com/SayedAlesawy/Videra-Storage/utils/requests"
	"github.com/julienschmidt/httprouter"
)

var scLogPrefix = "[Search-Controller]"

// SearchRequestHandler Handles client's search request
func (server *Server) SearchRequestHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	log.Println(scLogPrefix, "Received search request")

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
		start, startErr := time.Parse(requests.TimeStampLayout, r.Header.Get("StartTime"))
		end, endErr := time.Parse(requests.TimeStampLayout, r.Header.Get("EndTime"))

		if errors.IsError(startErr) || errors.IsError(endErr) {
			log.Println(scLogPrefix, r.RemoteAddr, err)
			requests.HandleRequestError(w, http.StatusBadRequest, err.Error())

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

	w.Header().Set("content-type", "application/json")
	w.Write(resp)
}

// retrieveClips A function to query the clips table for matching records
func retrieveClips(params ...interface{}) []namenode.Clip {
	log.Println("lol", params)
	var clips []namenode.Clip

	if len(params) == 1 {
		namenode.NodeInstance().DB.Connection.Where("tag = ?", params[0]).Find(&clips)
	} else {
		namenode.NodeInstance().DB.Connection.Where("tag = ? and start_time >= ? and end_time <= ?",
			params[0], params[1], params[2]).Find(&clips)
	}
	log.Println("lol", clips)
	return clips
}
