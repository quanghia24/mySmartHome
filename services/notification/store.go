package notification

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

func (s *Store) CreateNoti(noti types.NotiPayload) error {
	_, err := s.db.Exec("INSERT INTO noti (userId, ip) VALUES (?, ?)", noti.UserID, noti.Ip)
	return err
}

func (s *Store) GetNotiByUserId(userId int) (*types.NotiPayload, error) {
	row := s.db.QueryRow("SELECT userId, ip FROM noti WHERE userId = ? LIMIT 1", userId)
	
	noti := new(types.NotiPayload)
	err := row.Scan(&noti.UserID, &noti.Ip)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return noti, nil
}