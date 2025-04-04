package sensor

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/quanghia24/mySmartHome/services/auth"
	"github.com/quanghia24/mySmartHome/types"
	"github.com/quanghia24/mySmartHome/utils"
)

type Handler struct {
	store          types.SensorStore
	userStore      types.UserStore
	logSensorStore types.LogSensorStore
}

func NewHandler(store types.SensorStore, userStore types.UserStore, logSensorStore types.LogSensorStore) *Handler {
	return &Handler{
		store:          store,
		userStore:      userStore,
		logSensorStore: logSensorStore,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/sensors", auth.WithJWTAuth(h.createSensor, h.userStore)).Methods(http.MethodPost)
	router.HandleFunc("/sensors/{feed_id}", h.getSensorInfo).Methods(http.MethodGet)
}

func (h *Handler) getSensorInfo(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	feedId, _ := strconv.Atoi(params["feed_id"])

	sensor, err := h.store.GetSensorByFeedID(feedId)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("get sensor info:%s", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, sensor)
}

func (h *Handler) createSensor(w http.ResponseWriter, r *http.Request) {
	userId := auth.GetUserIDFromContext(r.Context())

	var payload types.CreateDevicePayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors))
		return
	}

	err := h.store.CreateSensor(types.Sensor{
		Title:   payload.Title,
		FeedKey: payload.FeedKey,
		FeedId:  payload.FeedID,
		Type:    payload.Type,
		UserID:  userId,
		RoomID:  payload.RoomID,
	})

	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	// err = h.logSensorStore.CreateLogSensor(types.LogSensor{
	// 	Type:     "creation",
	// 	Message:  fmt.Sprintf("%s got added to the system", payload.Title),
	// 	SensorID: payload.FeedID,
	// 	UserID:   userId,
	// 	Value:    "0",
	// })
	// if err != nil {
	// 	utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("log sensor creation error:%v", err))
	// 	return
	// }

	utils.WriteJSON(w, http.StatusCreated, nil)
}

func (h *Handler) StartSensorDataPolling() {
	ticker := time.NewTicker(15*60 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		sensors, err := h.store.GetAllSensor()
		if err != nil {
			fmt.Printf("error retrieving sensors: %v\n", err)
			continue
		}

		for _, sensor := range sensors {
			go h.updateSensorData(sensor)
		}
	}
}

func (h *Handler) updateSensorData(sensor types.Sensor) {
	// get adafruit value
	url := os.Getenv("AIOAPI") + sensor.FeedKey + "/data?limit=1"
	resp, err := http.Get(url)
	if err != nil {
		log.Println("sensor log:", err)
		return
	}
	defer resp.Body.Close()

	var payload []types.SensorDataPayload

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("sensor resp body:", err)
		return
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		log.Println("Error parsing JSON:", err)
		return
	}


	err = h.logSensorStore.CreateLogSensor(types.LogSensor{
		Type:     "data",
		Message:  fmt.Sprintf("%s data recored", payload[0].Value),
		SensorID: payload[0].FeedId,
		UserID:   sensor.UserID,
		Value:    payload[0].Value,
	})

	if err != nil {
		log.Println("sensor log create:", err)
		return
	}
}
