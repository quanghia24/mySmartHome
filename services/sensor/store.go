package sensor

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

func (s *Store) CreateSensor(sensor types.Sensor) error {
	_, err := s.db.Exec("INSERT INTO sensors (feedId, feedKey, title, type, userID, roomID) VALUES (?, ?, ?, ?, ?, ?)", sensor.FeedId, sensor.FeedKey, sensor.Title, sensor.Type, sensor.UserID, sensor.RoomID)
	return err
}

func (s *Store) GetSensorByFeedID(feedId int) (*types.DeviceDataPayload, error) {
	query := `
		SELECT s.feedId, s.feedKey, l.value, s.type, s.title, l.createdAt 
		FROM sensors s
		LEFT JOIN logs_sensor l ON s.feedId=l.sensorId
		WHERE s.feedId = ?
		ORDER BY l.createdAt DESC
		LIMIT 1
	`
	var sensorData types.DeviceDataPayload
	err := s.db.QueryRow(query, feedId).Scan(
		&sensorData.FeedID,
		&sensorData.FeedKey,
		&sensorData.Value,
		&sensorData.Type,
		&sensorData.Title,
		&sensorData.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no data found for feedId %d", feedId)
		}
		return nil, err
	}

	return &sensorData, nil
}

func (s *Store) GetAllSensor() ([]types.Sensor, error) {
	rows, err := s.db.Query("SELECT * FROM sensors")
	if err != nil {
		return nil, err
	}

	sensors := []types.Sensor{}

	for rows.Next() {
		s, err := scanIntoSensor(rows)
		if err != nil {
			return nil, err
		}

		sensors = append(sensors, *s)
	}
	return sensors, nil
}

func scanIntoSensor(row *sql.Rows) (*types.Sensor, error) {
	sensor := new(types.Sensor)

	err := row.Scan(
		&sensor.FeedId,
		&sensor.FeedKey,
		&sensor.Title,
		&sensor.Type,
		&sensor.UserID,
		&sensor.RoomID,
	)

	if err != nil {
		return nil, err
	}
	return sensor, nil

}
