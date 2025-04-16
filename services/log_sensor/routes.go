package log_sensor

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/quanghia24/mySmartHome/types"
	"github.com/quanghia24/mySmartHome/utils"
)

type Handler struct {
	store types.LogSensorStore
}

func NewHandler(store types.LogSensorStore) *Handler {
	return &Handler{
		store: store,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/logsensor", h.getSensorData).Methods(http.MethodGet)
	router.HandleFunc("/logsensor/{feed_id}/usage", h.getSensorUsage).Methods(http.MethodPost)
}

func (h *Handler) getSensorData(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusOK, "connection to sensor log data seemd to be ok")
}

func (h *Handler) getSensorUsage(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	feedId, err := strconv.Atoi(params["feed_id"])
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	var payload types.UsageRequestPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	// var logs []types.LogSensor

	logs, err := h.store.GetLogSensorsLast7HoursByFeedID(feedId, time.Now())
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}


	// if payload.End.IsZero() {
	// 	fmt.Println("get 7 days from now:")
	// 	utils.WriteJSON(w, http.StatusOK, details)

	// } else if payload.Start.IsZero() {
	// 	fmt.Println("without startTime")
	// 	fmt.Println("get closest 7 day from end timePoint")
	// 	utils.WriteJSON(w, http.StatusOK, logs)
	// } else {
	// 	fmt.Println("get interval")
	// 	utils.WriteJSON(w, http.StatusOK, logs)
	// }
	utils.WriteJSON(w, http.StatusOK, logs)
}



