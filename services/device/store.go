package device

import (
	"database/sql"

	"github.com/quanghia24/mySmartHome/types"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) CreateDevice(device types.Device) error {
	_, err := s.db.Exec("INSERT INTO devices (title, feedKey, feedId, userID, roomID) VALUES (?, ?, ?, ?, ?)", device.Title, device.FeedKey, device.FeedId, device.UserID, device.RoomID)
	return err
}

func (s *Store) GetDevicesByID(id int) ([]types.Device, error) {
	rows, err := s.db.Query("SELECT * FROM devices WHERE userID = ?", id)
	if err != nil {
		return nil, err
	}

	devices := []types.Device{}

	for rows.Next() {
		d, err := scanRowsIntoDevice(rows)
		if err != nil {
			return nil, err
		}

		devices = append(devices, *d)
	}
	return devices, nil
}

func (s *Store) GetDevicesInRoomID(id int) ([]types.Device, error) {
	rows, err := s.db.Query("SELECT * FROM devices WHERE roomID = ?", id)
	if err != nil {
		return nil, err
	}

	devices := []types.Device{}

	for rows.Next() {
		d, err := scanRowsIntoDevice(rows)
		if err != nil {
			return nil, err
		}

		devices = append(devices, *d)
	}
	return devices, nil
}

func scanRowsIntoDevice(rows *sql.Rows) (*types.Device, error) {
	device := new(types.Device)

	err := rows.Scan(
		&device.FeedId,
		&device.FeedKey,
		&device.Title,
		&device.UserID,
		&device.RoomID,
	)
	if err != nil {
		return nil, err
	}

	return device, nil
}