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

func (s *Store) DeleteRoom(roomId int, userId int) error {
	query := `DELETE FROM rooms WHERE rooms.id = ? AND rooms.userId = ?`
	_, err := s.db.Exec(query, roomId, userId)
	return err
}

func (s *Store) GetRoomsByUserID(userId int) ([]types.RoomInfoPayload, error) {
	query := `
		SELECT 
			r.id,
			r.title,

			COUNT(CASE WHEN d.type = 'fan' THEN 1 END) AS fanC,
			MAX(CASE WHEN d.type = 'fan' AND l.value > 0 THEN 1 ELSE 0 END) AS fanS,

			COUNT(CASE WHEN d.type = 'light' THEN 1 END) AS lightC,
			MAX(CASE WHEN d.type = 'light' AND l.value > 0 THEN 1 ELSE 0 END) AS lightS,

			COUNT(CASE WHEN d.type = 'door' THEN 1 END) AS doorC,
			MAX(CASE WHEN d.type = 'door' AND l.value > 0 THEN 1 ELSE 0 END) AS doorS,
      
      		(SELECT (COUNT(*)) FROM sensors s WHERE s.roomId = r.id) as sensorC 
      
    	FROM rooms r 
		LEFT JOIN devices d ON r.id = d.roomId
		LEFT JOIN logs l ON d.feedId = l.deviceId
		AND l.createdAt = (
			SELECT MAX(l2.createdAt) FROM logs l2 WHERE l2.deviceId = d.feedId
		)
		WHERE r.userId = ?
		GROUP BY r.id;
	`
	rows, err := s.db.Query(query, userId)
	if err != nil {
		return nil, err
	}

	rooms := []types.RoomInfoPayload{}
	for rows.Next() {
		r, err := scanRowsIntoRoom(rows)
		if err != nil {
			return nil, err
		}

		rooms = append(rooms, *r)
	}

	return rooms, nil
}

func (s *Store) GetDevicesByRoomId(roomId int) ([]int, error) {
	query := `
		SELECT feedId 
		FROM devices
		WHERE roomId = ?	
	`
	rows, err := s.db.Query(query, roomId)
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

func scanRowsIntoRoom(rows *sql.Rows) (*types.RoomInfoPayload, error) {
	room := new(types.RoomInfoPayload)

	err := rows.Scan(
		&room.ID,
		&room.Title,
		&room.FanCount,
		&room.FanStatus,
		&room.LightCount,
		&room.LightStatus,
		&room.DoorCount,
		&room.DoorStatus,
		&room.SensorCount,
	)
	if err != nil {
		return nil, err
	}
	return room, nil
}


func (s *Store) UpdateRoom(room types.Room) error {
	_, err := s.db.Exec(`
		UPDATE rooms
		SET title = ?, image = ?
		WHERE id = ?
	`, room.Title, room.Image, room.ID)
	return err
}