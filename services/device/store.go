package device

import (
	"database/sql"
	"fmt"

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
	_, err := s.db.Exec("INSERT INTO devices (feedId, feedKey, title, type, userID, roomID) VALUES (?, ?, ?, ?, ?, ?)", device.FeedId, device.FeedKey, device.Title, device.Type, device.UserID, device.RoomID)
	return err
}

func (s *Store) GetDevicesByUserID(userId int) ([]types.DeviceDataPayload, error) {
	query := `
		SELECT d.feedId, d.feedKey, l.value, d.type, d.title, l.createdAt
		FROM devices d
		LEFT JOIN logs l 
			ON d.feedId = l.deviceId
			AND l.createdAt = (
				SELECT MAX(l2.createdAt)
				FROM logs l2
				WHERE l2.deviceId = d.feedId
			)
		WHERE d.userID = ?;
	`

	rows, err := s.db.Query(query, userId)
	if err != nil {
		return nil, err
	}

	devices := []types.DeviceDataPayload{}

	for rows.Next() {
		d, err := scanRowsIntoDeviceDataPayload(rows)
		if err != nil {
			return nil, err
		}

		devices = append(devices, *d)
	}
	return devices, nil
}

func (s *Store) GetDevicesByFeedID(feedId int) (*types.DeviceDataPayload, error) {
	query := `
		SELECT d.feedId, d.feedKey, logs.value, d.type, d.title, logs.createdAt 
		FROM devices d
		LEFT JOIN logs ON d.feedId=logs.deviceId
		WHERE d.feedId = ?
		ORDER BY logs.createdAt DESC
		LIMIT 1
	`

	var deviceData types.DeviceDataPayload
	err := s.db.QueryRow(query, feedId).Scan(
		&deviceData.FeedID,
		&deviceData.FeedKey,
		&deviceData.Value,
		&deviceData.Type,
		&deviceData.Title,
		&deviceData.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no data found for feedId %d", feedId)
		}
		return nil, err
	}

	return &deviceData, nil
}

func (s *Store) GetDevicesInRoomID(roomId int) ([]types.DeviceDataPayload, error) {
	query := `
		SELECT d.feedId, d.feedKey, l.value, d.type, d.title, l.createdAt
		FROM devices d
		LEFT JOIN logs l 
			ON d.feedId = l.deviceId
			AND l.createdAt = (
				SELECT MAX(l2.createdAt)
				FROM logs l2
				WHERE l2.deviceId = d.feedId
			)
		WHERE d.roomId = ?;
	`
	rows, err := s.db.Query(query, roomId)
	if err != nil {
		return nil, err
	}

	devices := []types.DeviceDataPayload{}

	for rows.Next() {
		d, err := scanRowsIntoDeviceDataPayload(rows)
		if err != nil {
			return nil, err
		}

		devices = append(devices, *d)
	}
	return devices, nil
}

// func scanRowsIntoDevice(rows *sql.Rows) (*types.Device, error) {
// 	device := new(types.Device)

// 	err := rows.Scan(
// 		&device.FeedId,
// 		&device.FeedKey,
// 		&device.Title,
// 		&device.Type,
// 		&device.UserID,
// 		&device.RoomID,
// 	)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return device, nil
// }

func scanRowsIntoDeviceDataPayload(rows *sql.Rows) (*types.DeviceDataPayload, error) {
	device := new(types.DeviceDataPayload)

	err := rows.Scan(
		&device.FeedID,
		&device.FeedKey,
		&device.Value,
		&device.Type,
		&device.Title,
		&device.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return device, nil
}
