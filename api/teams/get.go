package teams

import (
	"encoding/json"
	"log"
	"net/http"
	"service-atlas/internal/customerrors"
	"strconv"

	"github.com/service-atlas/go-common/httphelpers"
)

func (c CallsHandler) GetTeam(rw http.ResponseWriter, r *http.Request) {
	id, ok := httphelpers.GetGuidFromRequestPath("id", r)
	if !ok {
		http.Error(rw, "Invalid team ID", http.StatusBadRequest)
		return
	}
	team, err := c.Repository.GetTeam(r.Context(), id)
	if err != nil {
		customerrors.HandleError(rw, err)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(rw).Encode(team)
	if err != nil {
		log.Println(err)
	}
}

func (c CallsHandler) GetTeams(rw http.ResponseWriter, r *http.Request) {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	if page < 1 {
		http.Error(rw, "page must be positive", http.StatusBadRequest)
		return
	}
	pageSize, err := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if err != nil {
		pageSize = 10
	}
	if pageSize < 1 || pageSize > 100 {
		http.Error(rw, "pageSize must be between 1 and 100", http.StatusBadRequest)
		return
	}
	teams, err := c.Repository.GetTeams(r.Context(), page, pageSize)
	if err != nil {
		customerrors.HandleError(rw, err)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(rw).Encode(teams)
	if err != nil {
		log.Println(err)
	}
}
