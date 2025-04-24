package plan

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/quanghia24/mySmartHome/utils"
)

type Handler struct {
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/plans/{feed_id}", h.makePlan).Methods(http.MethodPost)
}

func (h *Handler) makePlan(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	utils.WriteJSON(w, http.StatusOK, fmt.Sprintf("connection ok %s", params["feed_id"]))
}
