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
	deviceLog   types.LogDeviceStore
	sensorLog   types.LogSensorStore
	userStore   types.UserStore
	roomStore   types.RoomStore
	deviceStore types.DeviceStore
}

func NewHandler(deviceLog types.LogDeviceStore, sensorLog types.LogSensorStore, userStore types.UserStore, roomStore types.RoomStore, deviceStore types.DeviceStore) *Handler {
	return &Handler{
		deviceLog:   deviceLog,
		sensorLog:   sensorLog,
		userStore:   userStore,
		roomStore:   roomStore,
		deviceStore: deviceStore,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/statistic/device/{feed_id}", h.getDeviceStatistic).Methods(http.MethodPost)
	router.HandleFunc("/statistic/device/{feed_id}/total", h.getDeviceTotalStatistic).Methods(http.MethodPost)
	router.HandleFunc("/statistic/sensor/{feed_id}", h.getSensorStatistic).Methods(http.MethodPost)
	router.HandleFunc("/statistic/rooms", auth.WithJWTAuth(h.getRoomAllStatistic, h.userStore)).Methods(http.MethodPost)
	router.HandleFunc("/statistic/rooms/{room_id}", h.getStatisticByRoom).Methods(http.MethodPost)
	router.HandleFunc("/statistic/rooms/{room_id}/{device_type}", h.getRoomDeviceStatistic).Methods(http.MethodPost)
	router.HandleFunc("/statistic/rooms-electric", auth.WithJWTAuth(h.getElectricBills, h.userStore)).Methods(http.MethodGet)
	router.HandleFunc("/statistic/graph/{device_type}", auth.WithJWTAuth(h.getGraphicalStatistic, h.userStore)).Methods(http.MethodPost)
}

func (h *Handler) getGraphicalStatistic(w http.ResponseWriter, r *http.Request) {
	userId := auth.GetUserIDFromContext(r.Context())
	params := mux.Vars(r)
	mtype := params["device_type"]
	// get all rooms belong to user

	rooms, err := h.roomStore.GetRoomsByUserID(userId)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	var payload types.RequestStatisticDevicePayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("missing payload"))
		return
	}

	offvalue := "0"
	if mtype == "light" {
		offvalue = "#000000"
	}

	startDate := payload.Start.Truncate(24 * time.Hour) // today at 00:00
	endDate := payload.End.Add(24 * time.Hour)          // tomorrow 00:00

	if mtype == "all" {
		fresult := make(map[int]float64)
		lresult := make(map[int]float64)

		for _, room := range rooms {
			fresult[room.ID] = 0
			lresult[room.ID] = 0

			fans, err := h.deviceStore.GetDevicesByRoomIdAndType(room.ID, "fan")
			if err != nil {
				utils.WriteJSON(w, http.StatusInternalServerError, err)
				return
			}
			for _, deviceId := range fans {
				// Get logs of the device today
				logs, err := h.deviceLog.GetLogsByFeedIDBetween(deviceId, startDate, endDate)
				if err != nil {
					log.Println(err)
					continue
				}

				// Calculate total ON-time in hours
				totalHours := calculateOnTimeHours(logs, endDate, "0")

				// Apply energy multiplier
				fresult[room.ID] += totalHours
			}

			lights, err := h.deviceStore.GetDevicesByRoomIdAndType(room.ID, "light")
			if err != nil {
				utils.WriteJSON(w, http.StatusInternalServerError, err)
				return
			}
			for _, deviceId := range lights {
				// Get logs of the device today
				logs, err := h.deviceLog.GetLogsByFeedIDBetween(deviceId, startDate, endDate)
				if err != nil {
					log.Println(err)
					continue
				}

				// Calculate total ON-time in hours
				totalHours := calculateOnTimeHours(logs, endDate, "#000000")

				// Apply energy multiplier
				lresult[room.ID] += totalHours
			}
		}

		utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"fan":   fresult,
			"light": lresult,
		})
		return
	}

	result := make(map[int]float64)

	for _, room := range rooms {
		result[room.ID] = 0
		devices, err := h.deviceStore.GetDevicesByRoomIdAndType(room.ID, mtype)
		if err != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, err)
			return
		}

		for _, deviceId := range devices {
			// Get logs of the device today
			logs, err := h.deviceLog.GetLogsByFeedIDBetween(deviceId, startDate, endDate)
			if err != nil {
				log.Println(err)
				continue
			}

			// Calculate total ON-time in hours
			totalHours := calculateOnTimeHours(logs, endDate, offvalue)

			// Apply energy multiplier
			result[room.ID] += totalHours
		}
	}

	utils.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) getDeviceTotalStatistic(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	feed_id, _ := strconv.Atoi(params["feed_id"])

	var payload types.RequestStatisticDevicePayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("missing payload"))
		return
	}

	device, err := h.deviceStore.GetDevicesByFeedID(feed_id)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	offvalue := "0"
	if device.Type == "light" {
		offvalue = "#000000"
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
	dayHours := initDateList(payload.Start, payload.End) // date "YYYY-MM-DD" -> total running hours

	for _, log := range logs {
		createdAt := log.CreatedAt
		isOn := log.Value != offvalue // value 0 means off

		if isOn && lastOnTime == nil {
			// Device just turned on
			lastOnTime = &createdAt
		} else if !isOn && lastOnTime != nil {
			// Device just turned off -> calculate time
			hoursByDay := splitDurationByDay(*lastOnTime, createdAt)
			for day, hours := range hoursByDay {
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

	total := 0.0
	for _, value := range dayHours {
		total += value
	}

	utils.WriteJSON(w, http.StatusOK, map[string]float64{"total": total})
}

func (h *Handler) getRoomAllStatistic(w http.ResponseWriter, r *http.Request) {
	userId := auth.GetUserIDFromContext(r.Context())
	// get all rooms belong to user

	rooms, err := h.roomStore.GetRoomsByUserID(userId)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	var payload types.RequestStatisticDevicePayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	fanHours := initDateList(payload.Start, payload.End)
	lightHours := initDateList(payload.Start, payload.End)

	for _, room := range rooms {
		// get devices list
		fans, err := h.deviceStore.GetDevicesByRoomIdAndType(room.ID, "fan")
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}
		lights, err := h.deviceStore.GetDevicesByRoomIdAndType(room.ID, "light")
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		// get logs
		for _, fanId := range fans {
			logs, err := h.deviceLog.GetLogsByFeedIDBetween(fanId, payload.Start, payload.End)
			if err != nil {
				utils.WriteError(w, http.StatusInternalServerError, err)
				return
			}

			// Sort logs by createdAt (just in case DB doesn't guarantee)
			sort.Slice(logs, func(i, j int) bool {
				return logs[i].CreatedAt.Before(logs[j].CreatedAt)
			})

			var lastOnTime *time.Time

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
						fanHours[day] += hours
					}
					lastOnTime = nil
				}
			}

			// If still "on" at the end of log, close the period
			if lastOnTime != nil {
				endTime := payload.End
				hoursByDay := splitDurationByDay(*lastOnTime, endTime)
				for day, hours := range hoursByDay {
					fanHours[day] += hours
				}
			}
		}

		for _, lightId := range lights {
			logs, err := h.deviceLog.GetLogsByFeedIDBetween(lightId, payload.Start, payload.End)
			if err != nil {
				utils.WriteError(w, http.StatusInternalServerError, err)
				return
			}

			// Sort logs by createdAt (just in case DB doesn't guarantee)
			sort.Slice(logs, func(i, j int) bool {
				return logs[i].CreatedAt.Before(logs[j].CreatedAt)
			})

			var lastOnTime *time.Time

			for _, log := range logs {
				createdAt := log.CreatedAt
				isOn := log.Value != "#000000" // value 0 means off

				if isOn && lastOnTime == nil {
					// Device just turned on
					lastOnTime = &createdAt
				} else if !isOn && lastOnTime != nil {
					// Device just turned off -> calculate time
					hoursByDay := splitDurationByDay(*lastOnTime, createdAt)
					for day, hours := range hoursByDay {
						lightHours[day] += hours
					}
					lastOnTime = nil
				}
			}

			// If still "on" at the end of log, close the period
			if lastOnTime != nil {
				endTime := payload.End
				hoursByDay := splitDurationByDay(*lastOnTime, endTime)
				for day, hours := range hoursByDay {
					lightHours[day] += hours
				}
			}
		}

	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"fan":   fanHours,
		"light": lightHours,
	})
}

func (h *Handler) getStatisticByRoom(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	room_id, _ := strconv.Atoi(params["room_id"])

	var payload types.RequestStatisticDevicePayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	fanHours := initDateList(payload.Start, payload.End)
	lightHours := initDateList(payload.Start, payload.End)

	// get devices list
	fans, err := h.deviceStore.GetDevicesByRoomIdAndType(room_id, "fan")
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	lights, err := h.deviceStore.GetDevicesByRoomIdAndType(room_id, "light")
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	// get logs
	for _, fanId := range fans {
		logs, err := h.deviceLog.GetLogsByFeedIDBetween(fanId, payload.Start, payload.End)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		// Sort logs by createdAt (just in case DB doesn't guarantee)
		sort.Slice(logs, func(i, j int) bool {
			return logs[i].CreatedAt.Before(logs[j].CreatedAt)
		})

		var lastOnTime *time.Time

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
					fanHours[day] += hours
				}
				lastOnTime = nil
			}
		}

		// If still "on" at the end of log, close the period
		if lastOnTime != nil {
			endTime := payload.End
			hoursByDay := splitDurationByDay(*lastOnTime, endTime)
			for day, hours := range hoursByDay {
				fanHours[day] += hours
			}
		}
	}

	for _, lightId := range lights {
		logs, err := h.deviceLog.GetLogsByFeedIDBetween(lightId, payload.Start, payload.End)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		// Sort logs by createdAt (just in case DB doesn't guarantee)
		sort.Slice(logs, func(i, j int) bool {
			return logs[i].CreatedAt.Before(logs[j].CreatedAt)
		})

		var lastOnTime *time.Time

		for _, log := range logs {
			createdAt := log.CreatedAt
			isOn := log.Value != "#000000" // value 0 means off

			if isOn && lastOnTime == nil {
				// Device just turned on
				lastOnTime = &createdAt
			} else if !isOn && lastOnTime != nil {
				// Device just turned off -> calculate time
				hoursByDay := splitDurationByDay(*lastOnTime, createdAt)
				for day, hours := range hoursByDay {
					lightHours[day] += hours
				}
				lastOnTime = nil
			}
		}

		// If still "on" at the end of log, close the period
		if lastOnTime != nil {
			endTime := payload.End
			hoursByDay := splitDurationByDay(*lastOnTime, endTime)
			for day, hours := range hoursByDay {
				lightHours[day] += hours
			}
		}
	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"fan":   fanHours,
		"light": lightHours,
	})
}

func (h *Handler) getRoomDeviceStatistic(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	room_id, _ := strconv.Atoi(params["room_id"])
	mtype := params["device_type"]

	var payload types.RequestStatisticDevicePayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("missing payload"))
		return
	}

	var Total float64 = 0

	// getall device in room of type device_type
	devices, err := h.deviceStore.GetDevicesByRoomIdAndType(room_id, mtype)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, err)
		return
	}

	startDate := payload.Start.Truncate(24 * time.Hour) // today at 00:00
	endDate := payload.End.Add(24 * time.Hour)          // tomorrow 00:00

	// return total device type
	offvalue := "0"
	if mtype == "light" {
		offvalue = "#000000"
	}

	for _, deviceFeedId := range devices {
		// Get logs of the device today
		logs, err := h.deviceLog.GetLogsByFeedIDBetween(deviceFeedId, startDate, endDate)
		if err != nil {
			log.Println(err)
			continue
		}

		// Calculate total ON-time in hours
		totalHours := calculateOnTimeHours(logs, endDate, offvalue)

		// Apply energy multiplier
		Total += totalHours
	}

	utils.WriteJSON(w, http.StatusOK, map[string]float64{"total": Total})
}

func (h *Handler) getElectricBills(w http.ResponseWriter, r *http.Request) {
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

			// Apply energy multiplier
			if device.Type == "fan" {
				roomdata.Total += calculateOnTimeHours(logs, todayEnd, "0") * 3.0 // fan: 3kWh
			} else if device.Type == "light" {
				roomdata.Total += calculateOnTimeHours(logs, todayEnd, "#000000") * 2.0 // light: 2kWh
			}
		}

		roomStats = append(roomStats, roomdata)
	}

	utils.WriteJSON(w, http.StatusOK, roomStats)
}

func calculateOnTimeHours(logs []types.LogDevice, endOfDay time.Time, offvalue string) float64 {
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
		isOn := log.Value != offvalue

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
	result := initDateList(payload.Start, payload.End)
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

	device, err := h.deviceStore.GetDevicesByFeedID(feed_id)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	offvalue := "0"
	if device.Type == "light" {
		offvalue = "#000000"
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
	dayHours := initDateList(payload.Start, payload.End) // date "YYYY-MM-DD" -> total running hours

	for _, log := range logs {
		createdAt := log.CreatedAt
		isOn := log.Value != offvalue // value 0 means off

		if isOn && lastOnTime == nil {
			// Device just turned on
			lastOnTime = &createdAt
		} else if !isOn && lastOnTime != nil {
			// Device just turned off -> calculate time
			hoursByDay := splitDurationByDay(*lastOnTime, createdAt)
			for day, hours := range hoursByDay {
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

func initDateList(start, end time.Time) map[string]float64 {
	listHours := make(map[string]float64)
	curr := start
	for curr.Before(end) {
		// End of current day
		dayEnd := time.Date(curr.Year(), curr.Month(), curr.Day(), 23, 59, 59, 0, time.UTC)

		if dayEnd.After(end) {
			dayEnd = end
		}

		dayStr := curr.Format("2006-01-02")
		listHours[dayStr] = 0

		curr = dayEnd.Add(time.Second) // move to next day start
	}
	return listHours
}
