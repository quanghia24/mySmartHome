package room

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/quanghia24/mySmartHome/types"
)

func TestRoomServiceHandler(t *testing.T) {
	roomStore := &mockRoomStore{}
	userStore := &mockUserStore{}
	handler := NewHandler(roomStore, userStore)

	t.Run("Should failed if userID doesn't exist", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/rooms/999999999", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		router := mux.NewRouter()

		router.HandleFunc("/rooms/{userID}", handler.createRoom)
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expect status code %d, got %d", http.StatusBadRequest, rr.Code)
		}
	})
}

type mockRoomStore struct {
}

func (m *mockRoomStore) GetRoomsByUserID(id int) ([]types.RoomInfoPayload, error) {
	return nil, nil
}
func (m *mockRoomStore) CreateRoom(room types.Room) error {
	return nil
}

func (m *mockRoomStore) UpdateRoom(room types.Room) error {
	return nil
}

func (m *mockRoomStore) DeleteRoom(roomId int, userId int) error {
	return nil
}

type mockUserStore struct {
}

func (m *mockUserStore) GetUserByEmail(email string) (*types.User, error) {
	return nil, fmt.Errorf("user not fould")
}
func (m *mockUserStore) GetUserByID(id int) (*types.User, error) {
	return nil, nil
}
func (m *mockUserStore) CreateUser(user types.User) error {
	return nil
}

func (m *mockUserStore) UpdateProfile(user types.User) error {
	return nil
}