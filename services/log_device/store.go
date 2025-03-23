package log_device

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

func (s *Store) CreateLog(log types.Log) error {
	_, err := s.db.Exec("INSERT INTO logs (type, message, deviceID, userID, value) VALUES (?,?,?,?,?)", log.Type, log.Message, log.DeviceID, log.UserID, log.Value)
	return err

}

func (s *Store) GetLogsByID(id int) ([]types.Log, error) {
	rows, err := s.db.Query("SELECT * FROM logs WHERE userID = ?", id)
	if err != nil {
		return nil, err
	}

	logs := []types.Log{}

	for rows.Next() {
		l, err := scanRowIntoLog(rows)
		if err != nil {
			return nil, err
		}

		logs = append(logs, *l)
	}
	return logs, nil
}

func scanRowIntoLog(rows *sql.Rows) (*types.Log, error) {
	log := new(types.Log)

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
