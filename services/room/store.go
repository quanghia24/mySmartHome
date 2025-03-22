package room

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

func (s *Store) CreateRoom(room types.Room) error {
	_, err := s.db.Exec("INSERT INTO rooms (title, userID) VALUES (?, ?)", room.Title, room.UserID)
	return err
}

func (s *Store) GetRoomsByID(id int) ([]types.Room, error) {
	rows, err := s.db.Query("SELECT * FROM rooms WHERE userID = ?", id)
	if err != nil {
		return nil, err
	}

	rooms := []types.Room{}
	for rows.Next() {
		r, err := scanRowsIntoRoom(rows)
		if err != nil {
			return nil, err
		}

		rooms = append(rooms, *r)
	}

	return rooms, nil
}

func scanRowsIntoRoom(rows *sql.Rows) (*types.Room, error) {
	room := new(types.Room)

	err := rows.Scan(
		&room.ID,
		&room.Title,
		&room.UserID,
	)
	if err != nil {
		return nil, err
	}
	return room, nil
}
