package device

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	store     types.DeviceStore
	userStore types.UserStore
	roomStore types.RoomStore
	logStore  types.LogStore
}

func NewHandler(store types.DeviceStore, userStore types.UserStore, roomStore types.RoomStore, logStore types.LogStore) *Handler {
	return &Handler{
		store:     store,
		userStore: userStore,
		roomStore: roomStore,
		logStore:  logStore,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	// get
	router.HandleFunc("/devices", auth.WithJWTAuth(h.getAllDeviceBelongToID, h.userStore)).Methods(http.MethodGet)
	router.HandleFunc("/devices/{feed_id}/logs", h.getDeviceData).Methods(http.MethodGet)
	router.HandleFunc("/devices/{feed_id}", h.getDeviceInfo).Methods(http.MethodGet)
	router.HandleFunc("/devices/room/{roomID}", h.getAllDeviceInRoom).Methods(http.MethodGet)
	// post
	router.HandleFunc("/devices", auth.WithJWTAuth(h.createDevice, h.userStore)).Methods(http.MethodPost)
	router.HandleFunc("/devices/{feed_id}", auth.WithJWTAuth(h.addDeviceData, h.userStore)).Methods(http.MethodPost)
}

func (h *Handler) addDeviceData(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	feedId, err := strconv.Atoi(params["feed_id"])
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	device, err := h.store.GetDevicesByFeedID(feedId)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	url := os.Getenv("AIOAPI") + device.FeedKey + "/data"
	log.Println("adding data to", url)

	userId := auth.GetUserIDFromContext(r.Context())

	var payload types.DeviceDataPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	fmt.Println(payload)

	payload.CreatedAt = time.Now()

	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload %v", errors))
		return
	}

	// send request to adafruit server
	jsonData, err := json.Marshal(payload)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	apiKey := os.Getenv("AIOKey")
	if apiKey == "" {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("missing AIO Key"))
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-AIO-Key", apiKey)

	// make the request
	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	device.Value = payload.Value

	err = h.logStore.CreateLog(types.Log{
		Type:     "onoff",
		Message:  fmt.Sprintf("%s got value %s", device.FeedKey, device.Value),
		DeviceID: feedId,
		UserID:   userId,
		Value:    device.Value,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("log creation:%v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, device)
}

func (h *Handler) getDeviceData(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	feedId, err := strconv.Atoi(params["feed_id"])
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	logs, err := h.logStore.GetLogsByFeedID(feedId)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, logs)
}

func (h *Handler) getDeviceInfo(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	feedId, _ := strconv.Atoi(params["feed_id"])

	deviceData, err := h.store.GetDevicesByFeedID(feedId)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("get device info:%s", err))
		return
	}

	// var jsonResponse []types.DeviceDataPayload
	// if err := json.Unmarshal(responseData, &jsonResponse); err != nil {
	// 	utils.WriteError(w, http.StatusInternalServerError, err)
	// 	return
	// }

	// utils.WriteJSON(w, http.StatusOK, jsonResponse)
	utils.WriteJSON(w, http.StatusOK, deviceData)
}

func (h *Handler) getAllDeviceBelongToID(w http.ResponseWriter, r *http.Request) {
	userId := auth.GetUserIDFromContext(r.Context())

	_, err := h.userStore.GetUserByID(userId)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user %v doesn't exist", userId))
		return
	}
	// improve: check if room does exist

	devices, err := h.store.GetDevicesByUserID(userId)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	fmt.Println(userId, devices)

	utils.WriteJSON(w, http.StatusOK, devices)
}

func (h *Handler) getAllDeviceInRoom(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	roomId, _ := strconv.Atoi(params["roomID"])

	// improve: check if room does exist


	devices, err := h.store.GetDevicesInRoomID(roomId)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	response := map[string][]types.DeviceDataPayload{
		"fanList":    {},
		"lightList":  {},
		"doorList":   {},
		"sensorList": {},
	}

	for _, d := range devices {
		switch d.Type {
		case "fan":
			response["fanList"] = append(response["fanList"], d)
		case "light":
			response["lightList"] = append(response["lightList"], d)
		case "door":
			response["doorList"] = append(response["doorList"], d)
		case "sensor":
			response["sensorList"] = append(response["sensorList"], d)
		}
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) createDevice(w http.ResponseWriter, r *http.Request) {
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

	err := h.store.CreateDevice(types.Device{
		Title:   payload.Title,
		FeedKey: payload.FeedKey,
		FeedId:  payload.FeedID,
		Type:    payload.Type,
		UserID:  userId,
		RoomID:  payload.RoomID,
	})

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	value := "0" // Default value
	if payload.Type == "light" {
		value = "#000000"
	}

	err = h.logStore.CreateLog(types.Log{
		Type:     "creation",
		Message:  fmt.Sprintf("%s got added to the system", payload.Title),
		DeviceID: payload.FeedID,
		UserID:   userId,
		Value:    value,
	})

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("log creation error:%v", err))
		return
	}

	utils.WriteJSON(w, http.StatusCreated, nil)
}
