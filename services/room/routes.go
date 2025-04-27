package room

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

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
	router.HandleFunc("/rooms/{roomId}", auth.WithJWTAuth(h.deleteRoom, h.userStore)).Methods(http.MethodDelete)
	router.HandleFunc("/rooms/{roomId}", auth.WithJWTAuth(h.updateRoom, h.userStore)).Methods(http.MethodPut)
}

func (h *Handler) updateRoom(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	roomId, _ := strconv.Atoi(params["roomId"])
	userId := auth.GetUserIDFromContext(r.Context())


	var payload struct {
		Title string `json:"title"`
		Image string `json:"image"`
	}

	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return 
	}
	err := h.store.UpdateRoom(types.Room{
		ID: roomId,
		Title: payload.Title,
		UserID: userId,
		Image: payload.Image,

	})
    if err != nil {
        utils.WriteError(w, http.StatusInternalServerError, err)
        return
    }

    utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "room updated"})
}

func (h *Handler) deleteRoom(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	roomId, err := strconv.Atoi(params["roomId"])
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	userId := auth.GetUserIDFromContext(r.Context())

	err = h.store.DeleteRoom(roomId, userId)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, fmt.Sprintf("Room %v has been deleted", roomId))
}

func (h *Handler) getAllRoom(w http.ResponseWriter, r *http.Request) {
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
