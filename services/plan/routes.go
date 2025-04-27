package plan

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/quanghia24/mySmartHome/types"
	"github.com/quanghia24/mySmartHome/utils"
)

type Handler struct {
	store types.PlanStore
}

func NewHandler(store types.PlanStore) *Handler {
	return &Handler{
		store: store,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/plans/{feed_id}", h.getPlan).Methods(http.MethodGet)
	router.HandleFunc("/plans/{feed_id}", h.createPlan).Methods(http.MethodPost)
	router.HandleFunc("/plans/{feed_id}", h.removePlan).Methods(http.MethodDelete)
}

func (h *Handler) getPlan(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	feed_id, _ := strconv.Atoi(params["feed_id"])
	plan, err := h.store.GetPlansByFeedID(feed_id)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, plan)

}

func (h *Handler) createPlan(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	feed_id, _ := strconv.Atoi(params["feed_id"])
	var payload types.Plan
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	// remove existing plan -> only exist at a time
	err := h.store.RemovePlan(feed_id)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	payload.SensorID = feed_id

	fmt.Println(payload)
	if err := h.store.CreatePlan(payload); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, fmt.Sprintf("plan for %d created", feed_id))
}

func (h *Handler) removePlan(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	feed_id, _ := strconv.Atoi(params["feed_id"])

	// remove existing plan -> only exist at a time
	err := h.store.RemovePlan(feed_id)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, fmt.Sprintf("plan of %d has been removed", feed_id))
}