package log_device

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/quanghia24/mySmartHome/services/auth"
	"github.com/quanghia24/mySmartHome/types"
	"github.com/quanghia24/mySmartHome/utils"
)

type Handler struct {
	store     types.LogDeviceStore
	userStore types.UserStore
}

func NewHandler(store types.LogDeviceStore, userStore types.UserStore) *Handler {
	return &Handler{
		store:     store,
		userStore: userStore,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/logs", auth.WithJWTAuth(h.getAllDeviceBelongToID, h.userStore)).Methods(http.MethodGet)
	// router.HandleFunc("/logs/{feed_id}", auth.WithJWTAuth(h.getAllDeviceBelongToID, h.userStore)).Methods(http.MethodGet)
	router.HandleFunc("/logs/{feed_id}/usage", h.getDeviceUsage).Methods(http.MethodPost)
}



func (h *Handler) getDeviceUsage(w http.ResponseWriter, r *http.Request) {
	// get feedid
	// get payload
	// decide action
	// 1.[not sending end time] get all logs
	// 2.[not sending start time] get log 7 days ago from end
	// 3.[sending both] get interval
	// runningTimeObjects = {
	// 	type: string,
	// 	title: string,
	// 	data: timeADayObject[],
	// 	startDate: string,
	// 	endDate: string
	// };

	// export type timeADayObject = {
	// 	dayOfWeek: ('Mon' | 'Tue' | 'Wed' | 'Thu' | 'Fri' | 'Sat' | 'Sun'),
	// 	value: number (type: minute)
	// };

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

	var logs []types.LogDevice
	var details []types.TimeObject

	if payload.End.IsZero() {
		fmt.Println("get 7 days from now:")
		
		logs, err = h.store.GetLogsByFeedID7Days(feedId, time.Now())
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		details = calculateDailyOnTime(logs)

		utils.WriteJSON(w, http.StatusOK, details)

	} else if payload.Start.IsZero() {
		fmt.Println("without startTime")
		fmt.Println("get closest 7 day from end timePoint")

		logs, err = h.store.GetLogsByFeedID7Days(feedId, payload.End)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}


		fmt.Println(details)
		utils.WriteJSON(w, http.StatusOK, logs)
	} else {
		fmt.Println("get interval")
		logs, err = h.store.GetLogsByFeedIDBetween(feedId, payload.Start, payload.End)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}
		fmt.Println(details)
		utils.WriteJSON(w, http.StatusOK, logs)
	}
}

func (h *Handler) getAllDeviceBelongToID(w http.ResponseWriter, r *http.Request) {
	userId := auth.GetUserIDFromContext(r.Context())

	_, err := h.userStore.GetUserByID(userId)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("requested user doesn't exists"))
		return
	}

	logs, err := h.store.GetLogsByUserID(userId)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, logs)
}

func calculateDailyOnTime(logs []types.LogDevice) []types.TimeObject {
	if len(logs) == 0 {
		return nil
	}


	// Grouping by date: map[YYYY-MM-DD] -> onTimeMinutes
	onTimePerDay := make(map[string]int)

	// last seen ON timestamp (the log before an OFF)
	var prevTime *time.Time
	isOn := false

	for i := 0; i < len(logs); i++ {
		log := logs[i]
		createdAt := log.CreatedAt
		value := log.Value != "0"

		if value { // device is ON
			if !isOn {
				// start tracking from this point
				isOn = true
				prevTime = &createdAt
			}
		} else { // device is OFF
			if isOn && prevTime != nil {
				// track usage from prevTime to now
				duration := -prevTime.Sub(createdAt)
				minutes := int(duration.Minutes())

				if minutes > 0 {
					// assign the ON time to the day of the START TIME (when ON began)
					day := prevTime.Format("2006-01-02")
					onTimePerDay[day] += minutes
				}
			}
			isOn = false
			prevTime = nil
		}
	}

	// If it was still ON at the last known log (we don’t know when it turned off)
	// treat the last log timestamp as the last point
	if isOn && prevTime != nil {
		// use the oldest log's timestamp as the OFF time
		lastLog := logs[len(logs)-1].CreatedAt
		duration := prevTime.Sub(lastLog)
		minutes := int(duration.Minutes())

		if minutes > 0 {
			day := prevTime.Format("2006-01-02")
			onTimePerDay[day] += minutes
		}
	}

	// Create result for the last 7 days (based on latest log's day)
	latest := logs[len(logs)-1].CreatedAt.In(time.FixedZone("UTC+7", 7*60*60))
	startOfToday := time.Date(latest.Year(), latest.Month(), latest.Day(), 0, 0, 0, 0, latest.Location())


	results := []types.TimeObject{}
	for i := 6; i >= 0; i-- {
		date := startOfToday.AddDate(0, 0, -i)
		key := date.Format("2006-01-02")
		minutes := onTimePerDay[key]
		results = append(results, types.TimeObject{
			Date:  date,
			Value: fmt.Sprintf("%d", minutes),
		})
	}
 
	return results
}
