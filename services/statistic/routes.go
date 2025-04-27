package statistic

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/quanghia24/mySmartHome/services/auth"
	"github.com/quanghia24/mySmartHome/types"
	"github.com/quanghia24/mySmartHome/utils"
)

type Handler struct {
	deviceLog types.LogDeviceStore
	sensorLog types.LogSensorStore
	userStore types.UserStore
	roomStore types.RoomStore
	deviceStore types.DeviceStore
}

func NewHandler(deviceLog types.LogDeviceStore, sensorLog types.LogSensorStore, userStore types.UserStore, roomStore types.RoomStore, deviceStore types.DeviceStore) *Handler {
	return &Handler{
		deviceLog: deviceLog,
		sensorLog: sensorLog,
		userStore: userStore,
		roomStore: roomStore,
		deviceStore: deviceStore,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/statistic/device/{feed_id}", h.getDeviceStatistic).Methods(http.MethodPost)
	router.HandleFunc("/statistic/sensor/{feed_id}", h.getSensorStatistic).Methods(http.MethodPost)
	router.HandleFunc("/statistic/rooms", auth.WithJWTAuth(h.getRoomStatistic, h.userStore)).Methods(http.MethodGet)
	// router.HandleFunc("/statistic/room/{room_id}", h.getRoomStatistic).Methods(http.MethodPost)
}

func (h *Handler) getRoomStatistic(w http.ResponseWriter, r *http.Request) {
	// params := mux.Vars(r)
	// room_id, _ := strconv.Atoi(params["room_id"])

	userId := auth.GetUserIDFromContext(r.Context())
	// get all rooms belong to user

	rooms, err := h.roomStore.GetRoomsByUserID(userId)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	type roomData struct {
		Id    int
		Title string
		Total float64
	}

	roomStats := []roomData{}
	todayStart := time.Now().Truncate(24 * time.Hour) // today at 00:00
	todayEnd := todayStart.Add(24 * time.Hour)        // tomorrow 00:00
	for _, room := range rooms {
		var roomdata = roomData{
			Id:    room.ID,
			Title: room.Title,
			Total: 0,
		}
		devices, err := h.roomStore.GetDevicesByRoomId(room.ID)
		if err != nil {
			log.Println(err)
			continue
		}
		
		for _, deviceFeedId := range devices {
			// Fetch device info to know type
			device, err := h.deviceStore.GetDevicesByFeedID(deviceFeedId) 
			if err != nil {
				log.Println(err)
				continue
			}

			// Only care about fan and light
			if device.Type == "door" {
				continue
			}

			// Get logs of the device today
			logs, err := h.deviceLog.GetLogsByFeedIDBetween(deviceFeedId, todayStart, todayEnd)
			if err != nil {
				log.Println(err)
				continue
			}

			// Calculate total ON-time in hours
			totalHours := calculateOnTimeHours(logs, todayEnd)

			// Apply energy multiplier
			if device.Type == "fan" {
				roomdata.Total += totalHours * 3.0 // fan: 3kWh
			} else if device.Type == "light" {
				roomdata.Total += totalHours * 2.0 // light: 2kWh
			}
		}

		roomStats = append(roomStats, roomdata)
	}

	utils.WriteJSON(w, http.StatusOK, roomStats)
}

func calculateOnTimeHours(logs []types.LogDevice, endOfDay time.Time) float64 {
	if len(logs) == 0 {
		return 0
	}

	sort.Slice(logs, func(i, j int) bool {
		return logs[i].CreatedAt.Before(logs[j].CreatedAt)
	})

	var lastOnTime *time.Time
	var totalHours float64

	for _, log := range logs {
		createdAt := log.CreatedAt
		isOn := log.Value != "0"

		if isOn && lastOnTime == nil {
			lastOnTime = &createdAt
		} else if !isOn && lastOnTime != nil {
			totalHours += createdAt.Sub(*lastOnTime).Hours()
			lastOnTime = nil
		}
	}

	if lastOnTime != nil {
		totalHours += endOfDay.Sub(*lastOnTime).Hours()
	}

	return totalHours
}


func (h *Handler) getSensorStatistic(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	feed_id, _ := strconv.Atoi(params["feed_id"])

	var payload types.RequestStatisticDevicePayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("missing payload"))
		return
	}

	// get all log belong to

	logs, err := h.sensorLog.GetSensorsByFeedIDBetween(feed_id, payload.Start, payload.End)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	if len(logs) == 0 {
		utils.WriteJSON(w, http.StatusOK, map[string]float64{})
		return
	}

	// Group by date and calculate average
	type dateStat struct {
		total float64
		count int
	}
	dailyStats := make(map[string]*dateStat)

	for _, log := range logs {
		createdAt := log.CreatedAt.Format("2006-01-02") // Only keep date part
		value, err := strconv.ParseFloat(log.Value, 64)
		if err != nil {
			continue // skip if value cannot be parsed
		}

		if _, exists := dailyStats[createdAt]; !exists {
			dailyStats[createdAt] = &dateStat{}
		}
		dailyStats[createdAt].total += value
		dailyStats[createdAt].count++
	}

	// Build response
	result := make(map[string]float64)
	for date, stat := range dailyStats {
		result[date] = stat.total / float64(stat.count)
	}

	utils.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) getDeviceStatistic(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	feed_id, _ := strconv.Atoi(params["feed_id"])

	var payload types.RequestStatisticDevicePayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("missing payload"))
		return
	}

	// get all log
	logs, err := h.deviceLog.GetLogsByFeedIDBetween(feed_id, payload.Start, payload.End)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	if len(logs) == 0 {
		utils.WriteJSON(w, http.StatusOK, map[string]float64{})
		return
	}

	// Sort logs by createdAt (just in case DB doesn't guarantee)
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].CreatedAt.Before(logs[j].CreatedAt)
	})

	var lastOnTime *time.Time
	dayHours := make(map[string]float64) // date "YYYY-MM-DD" -> total running hours

	for _, log := range logs {
		createdAt := log.CreatedAt
		isOn := log.Value != "0" // value 0 means off

		if isOn && lastOnTime == nil {
			// Device just turned on
			lastOnTime = &createdAt
		} else if !isOn && lastOnTime != nil {
			// Device just turned off -> calculate time
			hoursByDay := splitDurationByDay(*lastOnTime, createdAt)
			for day, hours := range hoursByDay {
				fmt.Println("creasing hours")
				dayHours[day] += hours
			}
			lastOnTime = nil
		}
	}

	// If still "on" at the end of log, close the period
	if lastOnTime != nil {
		endTime := payload.End
		hoursByDay := splitDurationByDay(*lastOnTime, endTime)
		for day, hours := range hoursByDay {
			dayHours[day] += hours
		}
	}

	utils.WriteJSON(w, http.StatusOK, dayHours)
}

func splitDurationByDay(start, end time.Time) map[string]float64 {
	result := make(map[string]float64)

	curr := start
	for curr.Before(end) {
		// End of current day
		dayEnd := time.Date(curr.Year(), curr.Month(), curr.Day(), 23, 59, 59, 0, time.UTC)

		if dayEnd.After(end) {
			dayEnd = end
		}

		duration := dayEnd.Sub(curr)
		dayStr := curr.Format("2006-01-02")
		result[dayStr] += duration.Hours()

		curr = dayEnd.Add(time.Second) // move to next day start
	}

	return result
}
