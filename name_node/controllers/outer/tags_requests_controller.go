package outer

import (
	"encoding/json"
	"log"
	"net/http"

	namenode "github.com/SayedAlesawy/Videra-Storage/name_node"
	"github.com/SayedAlesawy/Videra-Storage/utils/errors"
	"github.com/SayedAlesawy/Videra-Storage/utils/requests"
	"github.com/julienschmidt/httprouter"
)

var tagsLogPrefix = "[Tags-Controller]"

// tagResponse Represents the tags response
type tagResponse struct {
	Tag string `json:"tag"`
}

// TagsRequestHandler Handles the dashboard tags request
func (server *Server) TagsRequestHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	log.Println(scLogPrefix, "Received tags request")

	w.Header().Set("content-type", "application/json")

	tags := retrieveTags()

	resp, err := json.Marshal(tags)
	if errors.IsError(err) {
		log.Println(scLogPrefix, r.RemoteAddr, err)
		requests.HandleRequestError(w, http.StatusInternalServerError, err.Error())

		return
	}

	w.Write(resp)
}

// retrieveTags A function to query the clips table for tags
func retrieveTags() []tagResponse {
	var tags []tagResponse

	namenode.NodeInstance().DB.Connection.Raw("select distinct(tag) from clips").Scan(&tags)

	return tags
}
