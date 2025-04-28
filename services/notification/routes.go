package notification

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/quanghia24/mySmartHome/services/auth"
	"github.com/quanghia24/mySmartHome/types"
	"github.com/quanghia24/mySmartHome/utils"
)

type Handler struct {
	store     types.NotiStore
	userStore types.UserStore
}

func NewHandler(store types.NotiStore, userStore types.UserStore) *Handler {
	return &Handler{
		store:     store,
		userStore: userStore,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/noti", auth.WithJWTAuth(h.handlerSendNoti, h.userStore)).Methods(http.MethodPost)
}

func (h *Handler) handlerSendNoti(w http.ResponseWriter, r *http.Request) {
	userId := auth.GetUserIDFromContext(r.Context())
	noti, err := h.store.GetNotiByUserId(userId)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return 
	}

	// var payload types.

	utils.WriteJSON(w, http.StatusOK, noti)
}
