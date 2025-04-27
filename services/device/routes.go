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

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/quanghia24/mySmartHome/services/auth"
	"github.com/quanghia24/mySmartHome/types"
	"github.com/quanghia24/mySmartHome/utils"
)

type Handler struct {
	store      types.DeviceStore
	userStore  types.UserStore
	roomStore  types.RoomStore
	logStore   types.LogDeviceStore
	doorStore  types.DoorStore
	mqttClient MQTT.Client
}

func NewHandler(store types.DeviceStore, userStore types.UserStore, roomStore types.RoomStore, logStore types.LogDeviceStore, doorStore types.DoorStore, mqttClient MQTT.Client) *Handler {
	return &Handler{
		store:      store,
		userStore:  userStore,
		roomStore:  roomStore,
		logStore:   logStore,
		doorStore:  doorStore,
		mqttClient: mqttClient,
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
	router.HandleFunc("/devices/{feed_id}/setpwd", h.setPassword).Methods(http.MethodPost)
	router.HandleFunc("/devices/{feed_id}/getpwd", h.getPassword).Methods(http.MethodGet)
	router.HandleFunc("/devices/{feed_id}/checkpwd", h.checkPassword).Methods(http.MethodPost)

	// delete
	router.HandleFunc("/devices/{feed_id}", auth.WithJWTAuth(h.deleteDevice, h.userStore)).Methods(http.MethodDelete)

}

func (h *Handler) deleteDevice(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	deviceId := params["feed_id"]
	userId := auth.GetUserIDFromContext(r.Context())

	err := h.store.DeleteDevice(deviceId, userId)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, fmt.Sprintf("Device %v has been deleted", deviceId))

}

func (h *Handler) setPassword(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	feedId, _ := strconv.Atoi(params["feed_id"])

	var payload struct {
		PWD string `json:"pwd"`
	}

	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	err := h.doorStore.CreatePassword(types.DoorPassword{
		FeedID: feedId,
		PWD:    payload.PWD,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, "pwd updated")
}

func (h *Handler) getPassword(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	feedId, _ := strconv.Atoi(params["feed_id"])

	var payload struct {
		PWD string `json:"pwd"`
	}

	pwd, err := h.doorStore.GetPassword(feedId)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	payload.PWD = pwd.PWD

	utils.WriteJSON(w, http.StatusOK, payload)
}

func (h *Handler) checkPassword(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	feedId, _ := strconv.Atoi(params["feed_id"])

	pwd, err := h.doorStore.GetPassword(feedId)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	if pwd.PWD == "" {
		utils.WriteJSON(w, http.StatusOK, "door unlocked")
		return
	}

	var payload struct {
		PWD string `json:"pwd"`
	}

	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if payload.PWD == pwd.PWD {
		utils.WriteJSON(w, http.StatusOK, "door unlocked")
		return
	}

	utils.WriteJSON(w, http.StatusUnauthorized, "wrong password")
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

	// userId := auth.GetUserIDFromContext(r.Context())

	var payload types.DeviceDataPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if device.Type == "fan" {
		if payload.Value == "1" {
			payload.Value = "50"
		} else if payload.Value == "2" {
			payload.Value = "75"
		} else if payload.Value == "3" {
			payload.Value = "100"
		} else {
			payload.Value = "0"
		}
	} else if device.Type == "door" {
		err := h.doorStore.CreatePassword(types.DoorPassword{
			FeedID: feedId,
			PWD:    "",
		})
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}
	}

	location, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		log.Fatalf("failed to load location: %v", err)
	}
	payload.CreatedAt = time.Now().In(location)

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

	// msg := ""
	// switch device.Type {
	// case "door":
	// 	if device.Value == "0" {
	// 		msg = fmt.Sprintf("[%s] got closed", device.Title)
	// 	} else {
	// 		msg = fmt.Sprintf("[%s] got opened", device.Title)
	// 	}
	// case "fan":
	// 	msg = fmt.Sprintf("[%s]'s set at level: %s", device.Title, device.Value)
	// case "light":
	// 	msg = fmt.Sprintf("[%s]'s set color: %s", device.Title, device.Value)
	// }

	// err = h.logStore.CreateLog(types.LogDevice{
	// 	Type:     "onoff",
	// 	Message:  msg,
	// 	DeviceID: feedId,
	// 	UserID:   userId,
	// 	Value:    device.Value,
	// })
	// if err != nil {
	// 	utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("log creation:%v", err))
	// 	return
	// }

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
		default:
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

	err = h.logStore.CreateLog(types.LogDevice{
		Type:     "creation",
		Message:  fmt.Sprintf("[%s] got added", payload.Title),
		DeviceID: payload.FeedID,
		UserID:   userId,
		Value:    value,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("log creation error:%v", err))
		return
	}
	
	if payload.Type == "door" {
		err := h.doorStore.CreatePassword(types.DoorPassword{
			FeedID: payload.FeedID,
			PWD:    "",
		})
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}
	}

	// mqtt
	topic := fmt.Sprintf("%s/feeds/%s", os.Getenv("AIOUSER"), payload.FeedKey)
	fmt.Println(topic)

	if token := h.mqttClient.Subscribe(topic, 0, func(client MQTT.Client, msg MQTT.Message) {
		fmt.Printf("New message on %s: %s\n", msg.Topic(), msg.Payload())

		// You can extract device ID by looking up payload.FeedKey or topic
		message := ""
		value := string(msg.Payload())
		switch payload.Type {
		case "door":
			if value == "0" {
				message = fmt.Sprintf("[%s] got closed", payload.Title)
			} else {
				message = fmt.Sprintf("[%s] got opened", payload.Title)
			}
		case "fan":
			message = fmt.Sprintf("[%s]'s set at level: %s", payload.Title, value)
		case "light":
			message = fmt.Sprintf("[%s]'s set color: %s", payload.Title, value)
		}

		err = h.logStore.CreateLog(types.LogDevice{
			Type:     "onoff",
			Message:  message,
			DeviceID: payload.FeedID,
			UserID:   userId,
			Value:    value,
		})
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("log creation:%v", err))
			return
		}
	}); token.Wait() && token.Error() != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("mqtt subscribe error: %v", token.Error()))
		return
	}

	utils.WriteJSON(w, http.StatusCreated, nil)
}
