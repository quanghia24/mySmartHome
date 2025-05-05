package user

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/quanghia24/mySmartHome/services/auth"
	"github.com/quanghia24/mySmartHome/types"
	"github.com/quanghia24/mySmartHome/utils"
)

// service
type Handler struct {
	store types.UserStore // store repository
	notiStore types.NotiStore
}

func NewHandler(store types.UserStore, notiStore types.NotiStore) *Handler {
	return &Handler{
		store: store,
		notiStore: notiStore,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/login", h.handleLogin).Methods("POST")
	router.HandleFunc("/register", h.handleRegister).Methods("POST")
	router.HandleFunc("/profile", auth.WithJWTAuth(h.handleGetProfile, h.store)).Methods("GET")
	router.HandleFunc("/profile", auth.WithJWTAuth(h.handleUpdateProfile, h.store)).Methods("PUT")

	router.HandleFunc("/logout", h.handleLogout).Methods("DELETE")
}

func (h *Handler) handleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	var payload types.User
	userId := auth.GetUserIDFromContext(r.Context())

	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}
	payload.ID = userId

	err := h.store.UpdateProfile(payload)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, "updated user profile")
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var payload types.LoginUserPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}
	//validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload %v", errors))
		return
	}

	// check exist
	u, err := h.store.GetUserByEmail(payload.Email)
	if err != nil { //user not exists
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("email or password was incorrect: %s", err))
		return
	}

	if !auth.ComparePasswords(u.Password, []byte(payload.Password)) {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("email or password was incorrect"))
		return
	}

	secret := []byte(os.Getenv("JWT_SECRET"))

	token, err := auth.CreateJWT(secret, u.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"token": token})
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	// receive payload
	// parse
	// validate
	var payload types.RegisterUserPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}
	//validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload %v", errors))
		return
	}

	// check exist, if doesn't then create new user
	_, err := h.store.GetUserByEmail(payload.Email)
	if err == nil { //user exists
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user with email %s already exists", payload.Email))
		return
	}

	// password needed to be hashed

	hashedPassword, err := auth.HashPassword(payload.Password)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	err = h.store.CreateUser(types.User{
		FirstName: payload.FirstName,
		LastName:  payload.LastName,
		Email:     payload.Email,
		Password:  hashedPassword,
	})

	if err != nil { //user exists
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}


	utils.WriteJSON(w, http.StatusCreated, map[string]string{"message": fmt.Sprintf("%s has been registered", payload.Email)})
}

func (h *Handler) handleGetProfile(w http.ResponseWriter, r *http.Request) {
	userId := auth.GetUserIDFromContext(r.Context())

	profile, err := h.store.GetUserByID(userId)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user doesn't exist in database"))
		return
	}
	fmt.Println(profile)

	utils.WriteJSON(w, http.StatusOK, profile)
}

func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "logout success"})
}
