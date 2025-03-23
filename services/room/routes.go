package room

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/quanghia24/mySmartHome/services/auth"
	"github.com/quanghia24/mySmartHome/types"
	"github.com/quanghia24/mySmartHome/utils"
)

type Handler struct {
	store     types.RoomStore
	userStore types.UserStore
}

func NewHandler(store types.RoomStore, userStore types.UserStore) *Handler {
	return &Handler{
		store:     store,
		userStore: userStore,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/rooms", auth.WithJWTAuth(h.getAllRoom, h.userStore)).Methods(http.MethodGet)
	router.HandleFunc("/rooms", auth.WithJWTAuth(h.createRoom, h.userStore)).Methods(http.MethodPost)
}

func (h *Handler) getAllRoom(w http.ResponseWriter, r *http.Request) {
	// params := mux.Vars(r)
	// id, _ := strconv.Atoi(params["userID"])
	id := auth.GetUserIDFromContext(r.Context())

	_, err := h.userStore.GetUserByID(id)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("requested user doesn't exists"))
		return
	}

	rooms, err := h.store.GetRoomsByUserID(id)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, rooms)
}


func (h *Handler) createRoom(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	log.Println("room added by", userID)

	var payload types.CreateRoomPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload %v", errors))
		return
	}

	err := h.store.CreateRoom(types.Room{
		Title:  payload.Title,
		UserID: userID,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	utils.WriteJSON(w, http.StatusCreated, nil)
}
