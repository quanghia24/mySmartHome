package log_sensor

import (
	"database/sql"
	"time"

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

func (s *Store) CreateLogSensor(log types.LogSensor) error {
	_, err := s.db.Exec("INSERT INTO logs_sensor (type, message, sensorID, userID, value) VALUES (?,?,?,?,?)", log.Type, log.Message, log.SensorID, log.UserID, log.Value)
	return err

}

func (s *Store) GetLogSensorsLast7HoursByFeedID(feedId int, end time.Time) ([]types.LogSensor, error) {
	start := end.Add(-7 * time.Hour) // 7 hours before the end time

	query := `
		SELECT * 
		FROM logs_sensor 
		WHERE sensorId = ?
		AND createdAt BETWEEN ? AND ?
		ORDER BY createdAt DESC
	`

	rows, err := s.db.Query(query, feedId, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := []types.LogSensor{}

	for rows.Next() {
		l, err := scanRowIntoLog(rows)
		if err != nil {
			return nil, err
		}
		logs = append(logs, *l)
	}

	return logs, nil
}

func (s *Store) GetLogSensorsByUserID(userId int) ([]types.LogSensor, error) {
	query := `
		SELECT * FROM logs_sensor WHERE userId = ?
		ORDER BY logs_sensor.createdAt DESC
	`
	rows, err := s.db.Query(query, userId)
	if err != nil {
		return nil, err
	}

	logs := []types.LogSensor{}

	for rows.Next() {
		l, err := scanRowIntoLog(rows)
		if err != nil {
			return nil, err
		}

		logs = append(logs, *l)
	}
	return logs, nil
}



func scanRowIntoLog(rows *sql.Rows) (*types.LogSensor, error) {
	log := new(types.LogSensor)

	err := rows.Scan(
		&log.ID,
		&log.Type,
		&log.Message,
		&log.SensorID,
		&log.UserID,
		&log.Value,
		&log.CreatedAt,
	)

	if err != nil {
		return nil, err
	}
	return log, nil
}
