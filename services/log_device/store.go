package log_device

import (
	"database/sql"
	"fmt"
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

func (s *Store) CreateLog(log types.LogDevice) error {
	_, err := s.db.Exec("INSERT INTO logs (type, message, deviceID, userID, value) VALUES (?,?,?,?,?)", log.Type, log.Message, log.DeviceID, log.UserID, log.Value)
	return err

}

func (s *Store) GetLogsByFeedID(feedId int) ([]types.LogDevice, error) {
	query := `
		SELECT id, type, message, deviceId, userId, value, CONVERT_TZ(logs.createdAt, '+00:00', '+07:00') AS createdAt_utc7 
		FROM logs 
		WHERE deviceId = ? 
		ORDER BY logs.createdAt DESC
	`

	rows, err := s.db.Query(query, feedId)
	if err != nil {
		return nil, err
	}

	logs := []types.LogDevice{}

	for rows.Next() {
		l, err := scanRowIntoLog(rows)
		if err != nil {
			return nil, err
		}

		logs = append(logs, *l)
	}
	return logs, nil
}

func (s *Store) GetLogsByFeedIDBetween(feedId int, start time.Time, end time.Time) ([]types.LogDevice, error) {
	query := `
		SELECT id, type, message, deviceId, userId, value, CONVERT_TZ(logs.createdAt, '+00:00', '+07:00') AS createdAt_utc7 
		FROM logs 
		WHERE deviceId = ? AND createdAt BETWEEN ? AND ?
		ORDER BY logs.createdAt DESC
	`

	rows, err := s.db.Query(query, feedId, start, end)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	logs := []types.LogDevice{}

	for rows.Next() {
		l, err := scanRowIntoLog(rows)
		if err != nil {
			return nil, err
		}

		logs = append(logs, *l)
	}
	return logs, nil
}

func (s *Store) GetLogsByFeedID7Days(feedId int, end time.Time) ([]types.LogDevice, error) {
	query := `
	SELECT 
		id, type, message, deviceId, userId, value, 
		CONVERT_TZ(logs.createdAt, '+00:00', '+07:00') AS createdAt_utc7
	FROM 
		logs
	WHERE 
		deviceId = ?
		AND createdAt >= (
			SELECT DATE_SUB(MAX(createdAt), INTERVAL 7 DAY)
			FROM logs
			WHERE deviceId = ?
		)
	ORDER BY 
		createdAt;
	`
	// SELECT id, type, message, deviceId, userId, value, CONVERT_TZ(logs.createdAt, '+00:00', '+07:00') AS createdAt_utc7 
	// FROM logs
	// WHERE deviceId = ? 
	// ORDER BY createdAt
		// AND createdAt BETWEEN 
		// 	DATE_SUB(DATE(?), INTERVAL 7 DAY)
		// 	AND ?
		// ORDER BY CONVERT_TZ(createdAt, '+00:00', '+07:00');

	fmt.Println(end)
	rows, err := s.db.Query(query, feedId, feedId)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	logs := []types.LogDevice{}

	for rows.Next() {
		l, err := scanRowIntoLog(rows)
		if err != nil {
			return nil, err
		}

		logs = append(logs, *l)
	}
	return logs, nil
}

func (s *Store) GetLogsByUserID(userId int) ([]types.LogDevice, error) {
	query := `
		SELECT id, type, message, deviceId, userId, value, CONVERT_TZ(logs.createdAt, '+00:00', '+07:00') AS createdAt_utc7 
		FROM logs 
		WHERE userId = ?
		ORDER BY logs.createdAt DESC
	`
	rows, err := s.db.Query(query, userId)
	if err != nil {
		return nil, err
	}

	logs := []types.LogDevice{}

	for rows.Next() {
		l, err := scanRowIntoLog(rows)
		if err != nil {
			return nil, err
		}

		logs = append(logs, *l)
	}
	return logs, nil
}

func scanRowIntoLog(rows *sql.Rows) (*types.LogDevice, error) {
	log := new(types.LogDevice)

	err := rows.Scan(
		&log.ID,
		&log.Type,
		&log.Message,
		&log.DeviceID,
		&log.UserID,
		&log.Value,
		&log.CreatedAt,
	)

	if err != nil {
		return nil, err
	}
	return log, nil
}
