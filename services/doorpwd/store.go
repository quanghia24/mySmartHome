package doorpwd

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

func (s *Store) CreatePassword(pwd types.DoorPassword) error {
	_, err := s.db.Exec(`
		INSERT INTO door_pwd (feedId, pwd)
		VALUES (?, ?)
		ON DUPLICATE KEY UPDATE
		pwd = VALUES(pwd)
	`, pwd.FeedID, pwd.PWD)
	return err
}

func (s *Store) GetPassword(feedId int) (*types.DoorPassword, error) {
	query := `
		SELECT * 
		FROM door_pwd 
		WHERE door_pwd.feedId = ?
		ORDER BY door_pwd.createdAt DESC
		LIMIT 1
	`

	var doorData types.DoorPassword
	err := s.db.QueryRow(query, feedId).Scan(
		&doorData.ID,
		&doorData.FeedID,
		&doorData.PWD,
		&doorData.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no data found for feedId %d", feedId)
		}
		return nil, err
	}

	return &doorData, nil
}
