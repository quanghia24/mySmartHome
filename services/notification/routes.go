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
	router.HandleFunc("/noti-ip", auth.WithJWTAuth(h.handleGetNotiIp, h.userStore)).Methods(http.MethodGet)
	router.HandleFunc("/noti-ip", auth.WithJWTAuth(h.handleCreateNotiIp, h.userStore)).Methods(http.MethodPost)

	router.HandleFunc("/noti", auth.WithJWTAuth(h.handleGetNoti, h.userStore)).Methods(http.MethodGet)
	router.HandleFunc("/noti", auth.WithJWTAuth(h.handleCreateNoti, h.userStore)).Methods(http.MethodPost)
}

func (h *Handler) handleGetNoti(w http.ResponseWriter, r *http.Request) {
	userId := auth.GetUserIDFromContext(r.Context())
	notis, err := h.store.GetNotiByUserId(userId)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return 
	}
	utils.WriteJSON(w, http.StatusOK, notis)
}

func (h *Handler) handleCreateNoti(w http.ResponseWriter, r *http.Request) {
	userId := auth.GetUserIDFromContext(r.Context())
	var payload types.NotiPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}
	ip, err := h.store.GetNotiIpByUserId(userId)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	payload.Ip = ip.Ip
	payload.UserID = userId

	if err := h.store.CreateNoti(payload); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return	
	}

	utils.WriteJSON(w, http.StatusOK, payload)
}

func (h *Handler) handleGetNotiIp(w http.ResponseWriter, r *http.Request) {
	userId := auth.GetUserIDFromContext(r.Context())
	noti, err := h.store.GetNotiIpByUserId(userId)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	// var payload types.
	utils.WriteJSON(w, http.StatusOK, noti)
}

func (h *Handler) handleCreateNotiIp(w http.ResponseWriter, r *http.Request) {
	userId := auth.GetUserIDFromContext(r.Context())

	var payload types.NotiIpPayload

	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	err := h.store.CreateNotiIp(types.NotiIpPayload{
		UserID: userId,
		Ip:     payload.Ip,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	// var payload types.
	utils.WriteJSON(w, http.StatusOK, "ip updated")
}

