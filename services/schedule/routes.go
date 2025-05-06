package schedule

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/quanghia24/mySmartHome/services/auth"
	"github.com/quanghia24/mySmartHome/types"
	"github.com/quanghia24/mySmartHome/utils"
	"github.com/robfig/cron/v3"
)

type Handler struct {
	store       types.ScheduleStore
	deviceStore types.DeviceStore
	logStore    types.LogDeviceStore
	doorStore   types.DoorStore
	userStore   types.UserStore
}

func NewHandler(store types.ScheduleStore, deviceStore types.DeviceStore, logStore types.LogDeviceStore, doorStore types.DoorStore, userStore types.UserStore) *Handler {
	return &Handler{
		store:       store,
		deviceStore: deviceStore,
		logStore:    logStore,
		doorStore:   doorStore,
		userStore:   userStore,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/schedules", auth.WithJWTAuth(h.createSchedule, h.userStore)).Methods(http.MethodPost)
	router.HandleFunc("/schedules/active", h.getAllActiveSchedule).Methods(http.MethodGet)
	router.HandleFunc("/schedules/{feed_id}", h.getDeviceScheduleByFeedId).Methods(http.MethodGet)
	router.HandleFunc("/schedules/{id}", h.updateDeviceSchedule).Methods(http.MethodPatch)
	router.HandleFunc("/schedules/{id}", h.removeDeviceSchedule).Methods(http.MethodDelete)

}

func (h *Handler) removeDeviceSchedule(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])
	if err := h.store.RemoveSchedule(id); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, fmt.Sprintf("schedule %d has been removed", id))
}

func (h *Handler) createSchedule(w http.ResponseWriter, r *http.Request) {
	var payload types.Schedule
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	userId := auth.GetUserIDFromContext(r.Context())
	payload.UserID = userId

	err := h.store.CreateSchedule(payload)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, payload)
}

func (h *Handler) getDeviceScheduleByFeedId(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	schedules, err := h.store.GetScheduleByFeedId(params["feed_id"])
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, schedules)
}

func (h *Handler) updateDeviceSchedule(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	var payload types.Schedule
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	payload.ID = id

	// if err := h.store.UpdateSchedule(payload); err != nil {
	// 	utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to update schedule: %v", err))
	// 	return
	// }

	// updated, err := h.store.GetScheduleByID(id)
	// if err != nil {
	// 	utils.WriteError(w, http.StatusInternalServerError,  fmt.Errorf("failed to fetch updated schedule: %v", err))
	// 	return
	// }

	utils.WriteJSON(w, http.StatusOK, payload)
}

func (h *Handler) getAllActiveSchedule(w http.ResponseWriter, r *http.Request) {
	schedules, err := h.store.GetAllActiveSchedule()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, schedules)

}

func (h *Handler) StartSchedule() {
	c := cron.New(cron.WithSeconds())
	c.AddFunc("0 * * * * *", func() {
		h.checkAndRunSchedules()
	})

	c.Start()
}

func (h *Handler) checkAndRunSchedules() {
	// fmt.Println("checking schedule")
	schedules, err := h.store.GetAllActiveSchedule()
	if err != nil {
		fmt.Printf("error at checking schedule: %v\n", err)
		return
	}


	for _, s := range schedules {
		loc, err := time.LoadLocation(s.Timezone)
		if err != nil {
			fmt.Println("Invalid timezone:", s.Timezone)
			continue
		}

		now := time.Now().In(loc)
		nowStr := now.Format("15:04")   // current time in HH:MM
		schedStr := s.ScheduledTime[:5] // e.g., "07:30:00" → "07:30"

		day := now.Weekday().String()[:3] // "Monday" → "Mon"

		if nowStr == schedStr && h.containsDay(s.RepeatDays, day) {
			fmt.Println("CREATEEEEEEEEEEEEE LOGGGGGGGGG")
			h.CreateDeviceData(s.DeviceID, s.Action, s.UserID)
		} else {
			fmt.Println("Same day:", h.containsDay(s.RepeatDays, day))
			fmt.Println(nowStr, schedStr)
		}
	}
}

func (h *Handler) containsDay(dayList string, day string) bool {
	// Simple check: look for "Mon", "Tue", etc.
	days := map[string]bool{}
	for _, d := range h.split(dayList) {
		days[d] = true
	}
	return days[day]
}

func (h *Handler) split(csv string) []string {
	var out []string
	for _, p := range [7]string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"} {
		if h.contains(csv, p) {
			out = append(out, p)
		}
	}
	return out
}

func (h *Handler) contains(haystack, needle string) bool {
	return len(haystack) >= 3 && (haystack == needle ||
		len(haystack) > 3 && (haystack[:3] == needle ||
			haystack[len(haystack)-3:] == needle ||
			haystack[1:4] == needle))
}

// !!!!!!!!!!



func (h *Handler) CreateDeviceData(feedId int, value string, userId int) error {
	device, err := h.deviceStore.GetDevicesByFeedID(feedId)
	if err != nil {
		return err
	}

	url := os.Getenv("AIOAPI") + device.FeedKey + "/data"
	log.Println("adding data to", url)

	// convert value for fan levels
	if device.Type == "fan" {
		switch value {
		case "1":
			value = "50"
		case "2":
			value = "75"
		case "3":
			value = "100"
		default:
			value = "0"
		}
	}

	payload := types.DeviceDataPayload{
		Value:     value,
		CreatedAt: time.Now().In(time.FixedZone("UTC+7", 7*3600)),
	}

	// send request to Adafruit
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	apiKey := os.Getenv("AIOKey")
	if apiKey == "" {
		return fmt.Errorf("missing AIO Key")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-AIO-Key", apiKey)

	client := &http.Client{}
	if _, err := client.Do(req); err != nil {
		return err
	}
	
	return nil
}
 