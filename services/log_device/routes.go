package log_device

import (
	"fmt"
	"net/http"

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
