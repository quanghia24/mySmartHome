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

func (s *Store) GetAllDevices() ([]types.AllDeviceDataPayload, error) {
	dquery := `
		SELECT d.feedId, d.feedKey, l.value, d.type, d.title, d.userId, l.createdAt
		FROM devices d
		LEFT JOIN logs l 
			ON d.feedId = l.deviceId
			AND l.createdAt = (
				SELECT MAX(l2.createdAt)
				FROM logs l2
				WHERE l2.deviceId = d.feedId
			)
	`

	drows, err := s.db.Query(dquery)
	if err != nil {
		return nil, err
	}

	
	devices := []types.AllDeviceDataPayload{}


	for drows.Next() {
		d, err := scanRowsIntoAllDeviceDataPayload(drows)
		if err != nil {
			return nil, err
		}

		devices = append(devices, *d)
	}

	return devices, nil
}

func (s *Store) GetDevicesByRoomIdAndType(roomId int, mtype string) ([]int, error){
	query := `
		SELECT feedId 
		FROM devices
		WHERE roomId = ? AND type = ?	
	`
	rows, err := s.db.Query(query, roomId, mtype)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feedIds []int

	for rows.Next() {
		var feedId int
		if err := rows.Scan(&feedId); err != nil {
			return nil, err
		}
		feedIds = append(feedIds, feedId)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return feedIds, nil
}

func (s *Store) GetDevicesByUserID(userId int) ([]types.DeviceDataPayload, error) {
	dquery := `
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

	drows, err := s.db.Query(dquery, userId)
	if err != nil {
		return nil, err
	}

	squery := `
		SELECT d.feedId, d.feedKey, l.value, d.type, d.title, l.createdAt
		FROM sensors d
		LEFT JOIN logs_sensor l 
			ON d.feedId = l.sensorId
			AND l.createdAt = (
				SELECT MAX(l2.createdAt)
				FROM logs_sensor l2
				WHERE l2.sensorId = d.feedId
			)
		WHERE d.userID = ?;
	`

	srows, err := s.db.Query(squery, userId)
	if err != nil {
		return nil, err
	}



	devices := []types.DeviceDataPayload{}




	for drows.Next() {
		d, err := scanRowsIntoDeviceDataPayload(drows)
		if err != nil {
			return nil, err
		}

		devices = append(devices, *d)
	}

	for srows.Next() {
		d, err := scanRowsIntoDeviceDataPayload(srows)
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
	dquery := `
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

	squery := `
		SELECT d.feedId, d.feedKey, l.value, d.type, d.title, l.createdAt
		FROM sensors d
		LEFT JOIN logs_sensor l 
			ON d.feedId = l.sensorId
			AND l.createdAt = (
				SELECT MAX(l2.createdAt)
				FROM logs_sensor l2
				WHERE l2.sensorId = d.feedId
			)
		WHERE d.roomId = ?;
	`

	drows, err := s.db.Query(dquery, roomId)
	if err != nil {
		return nil, err
	}

	srows, err := s.db.Query(squery, roomId)
	if err != nil {
		return nil, err
	}

	devices := []types.DeviceDataPayload{}

	for drows.Next() {
		d, err := scanRowsIntoDeviceDataPayload(drows)
		if err != nil {
			return nil, err
		}

		devices = append(devices, *d)
	}
	for srows.Next() {
		d, err := scanRowsIntoDeviceDataPayload(srows)
		if err != nil {
			return nil, err
		}

		devices = append(devices, *d)
	}

	return devices, nil
}

func (s *Store) DeleteDevice(deviceId string, userId int) error {
	query := `
		DELETE FROM devices
		WHERE feedId = ? AND userId = ?`
	_, err := s.db.Exec(query, deviceId, userId)
	return err	
}

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

func scanRowsIntoAllDeviceDataPayload(rows *sql.Rows) (*types.AllDeviceDataPayload, error) {
	device := new(types.AllDeviceDataPayload)

	err := rows.Scan(
		&device.FeedID,
		&device.FeedKey,
		&device.Value,
		&device.Type,
		&device.Title,
		&device.UserID,
		&device.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return device, nil
}

