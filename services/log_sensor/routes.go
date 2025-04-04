package log_sensor

import "github.com/quanghia24/mySmartHome/types"

type Handler struct {
	store types.LogSensorStore
}

func NewHandler(store types.LogSensorStore) *Handler {
	return &Handler{
		store: store,
	}
}
