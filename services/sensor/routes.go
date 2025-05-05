package sensor

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/quanghia24/mySmartHome/services/auth"
	"github.com/quanghia24/mySmartHome/types"
	"github.com/quanghia24/mySmartHome/utils"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

type Handler struct {
	store          types.SensorStore
	userStore      types.UserStore
	logSensorStore types.LogSensorStore
	planStore      types.PlanStore
	mqttClient     MQTT.Client
}

func NewHandler(store types.SensorStore, userStore types.UserStore, logSensorStore types.LogSensorStore, planStore types.PlanStore, mqttClient MQTT.Client) *Handler {
	return &Handler{
		store:          store,
		userStore:      userStore,
		logSensorStore: logSensorStore,
		planStore:      planStore,
		mqttClient:     mqttClient,
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
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	err = h.logSensorStore.CreateLogSensor(types.LogSensor{
		Type:     "creation",
		Message:  fmt.Sprintf("[%s] got added", payload.Title),
		SensorID: payload.FeedID,
		UserID:   userId,
		Value: "0",
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return 
	}
	

	// mqtt
	topic := fmt.Sprintf("%s/feeds/%s", os.Getenv("AIOUSER"), payload.FeedKey)
	fmt.Println(topic)

	if token := h.mqttClient.Subscribe(topic, 0, func(client MQTT.Client, msg MQTT.Message) {
		fmt.Printf("New message on %s: %s\n", msg.Topic(), msg.Payload())
		// check for plan

		f, _ := strconv.ParseFloat(string(msg.Payload()), 32)

		// Round to 1 decimal place
		value := math.Round(f*10) / 10

		// check for plan -> threshold
		// fmt.Println("Check threshold for", payload.FeedID, "with value of", value)
		plan, _ := h.planStore.GetPlansByFeedID(payload.FeedID)
		// if err != nil {
		// 	fmt.Println("Failed to get plans:", err)
		// }
		if plan != nil {
			fmt.Println(*plan)
			if plan.Lower != "" {
				lower, _ := strconv.ParseFloat(plan.Lower, 32)
				if lower > value {
					fmt.Println("WARNING!!! lower")
					err = h.logSensorStore.CreateLogSensor(types.LogSensor{
						Type:     "warning",
						Message:  fmt.Sprintf("%f below the %f lower bound", value, lower),
						SensorID: payload.FeedID,
						UserID:   userId,
						Value:    string(msg.Payload()),
					})

					if err != nil {
						log.Println("sensor log create:", err)
					}
				}
			}
			if plan.Upper != "" {
				upper, _ := strconv.ParseFloat(plan.Upper, 32)
				if upper < value {
					fmt.Println("WARNING!!! upper")
					err = h.logSensorStore.CreateLogSensor(types.LogSensor{
						Type:     "warning",
						Message:  fmt.Sprintf("%f exceed the %f upper bound", value, upper),
						SensorID: payload.FeedID,
						UserID:   userId,
						Value:    string(msg.Payload()),
					})

					if err != nil {
						log.Println("sensor log create:", err)
					}
				}
			}
		}

	}); token.Wait() && token.Error() != nil {
		fmt.Println("Failed to subscribe:", token.Error())
	}

	utils.WriteJSON(w, http.StatusCreated, nil)
}

func (h *Handler) StartSensorDataPolling() {
	ticker := time.NewTicker(15 * 60 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		fmt.Println("retrieve sensor data")
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
