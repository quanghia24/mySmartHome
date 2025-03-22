package device

import (
	"bytes"
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
	store     types.DeviceStore
	userStore types.UserStore
	roomStore types.RoomStore
}

func NewHandler(store types.DeviceStore, userStore types.UserStore, roomStore types.RoomStore) *Handler {
	return &Handler{
		store:     store,
		userStore: userStore,
		roomStore: roomStore,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	// get
	router.HandleFunc("/devices", auth.WithJWTAuth(h.getAllDeviceBelongToID, h.userStore)).Methods(http.MethodGet)
	router.HandleFunc("/devices/{feed_key}/data", h.getDeviceData).Methods(http.MethodGet)
	router.HandleFunc("/devices/{feed_key}", h.getCurrentStatus).Methods(http.MethodGet)
	router.HandleFunc("/devices/room/{roomID}", h.getAllDeviceInRoom).Methods(http.MethodGet)
	// post
	router.HandleFunc("/devices", auth.WithJWTAuth(h.createDevice, h.userStore)).Methods(http.MethodPost)
	router.HandleFunc("/devices/{feed_key}", h.addDeviceData).Methods(http.MethodPost)
}

func (h *Handler) addDeviceData(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	url := os.Getenv("AIOAPI") + params["feed_key"] + "/data"
	log.Println("adding data to", url)

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
	resp, err := client.Do(req)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	defer resp.Body.Close()

	// read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	var jsonResponse types.DeviceDataPayload
	if err := json.Unmarshal(respBody, &jsonResponse); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, jsonResponse)
}

func (h *Handler) getDeviceData(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	url := os.Getenv("AIOAPI") + params["feed_key"] + "/data"
	log.Println("calling", url)

	res, err := http.Get(url)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}
	responseData, err := io.ReadAll(res.Body)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	var jsonResponse []types.DeviceDataPayload
	if err := json.Unmarshal(responseData, &jsonResponse); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, jsonResponse)
}

func (h *Handler) getCurrentStatus(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	url := os.Getenv("AIOAPI") + params["feed_key"] + "/data?limit=1"
	log.Println("calling", url)

	res,err := http.Get(url)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	responseData, err := io.ReadAll(res.Body)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	var jsonResponse [] types.DeviceDataPayload
	if err := json.Unmarshal(responseData, &jsonResponse); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, jsonResponse)
}

func (h *Handler) getAllDeviceBelongToID(w http.ResponseWriter, r *http.Request) {
	id := auth.GetUserIDFromContext(r.Context())

	_, err := h.userStore.GetUserByID(id)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user %v doesn't exist", id))
		return
	}
	// improve: check if room does exist

	devices, err := h.store.GetDevicesByID(id)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, devices)
}

func (h *Handler) getAllDeviceInRoom(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["roomID"])

	_, err := h.userStore.GetUserByID(id)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("no device found in room %v", id))
		return
	}
	// improve: check if room does exist

	devices, err := h.store.GetDevicesInRoomID(id)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, devices)
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
		UserID:  userId,
		RoomID:  payload.RoomID,
	})

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, nil)
}
