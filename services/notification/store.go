package notification

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

func (s *Store) CreateNotiIp(noti types.NotiIpPayload) error {
	_, err := s.db.Exec("INSERT INTO `noti-ip` (userId, ip) VALUES (?, ?) ON DUPLICATE KEY UPDATE ip = VALUES(ip)", noti.UserID, noti.Ip)
	return err
}

func (s *Store) GetNotiIpByUserId(userId int) (*types.NotiIpPayload, error) {
	row := s.db.QueryRow("SELECT userId, ip FROM `noti-ip` WHERE userId = ? LIMIT 1", userId)

	noti := new(types.NotiIpPayload)
	err := row.Scan(&noti.UserID, &noti.Ip)
	if err == sql.ErrNoRows {
		fmt.Println("no rows have been found")
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return noti, nil
}

func (s *Store) CreateNoti(noti types.NotiPayload) error {
	_, err := s.db.Exec("insert into noti (userId, ip, message) values (?, ?, ?)", noti.UserID, noti.Ip, noti.Message)
	return err
}

func (s *Store) GetNotiByUserId(userId int) ([]types.NotiPayload, error) {
	rows, err := s.db.Query("select * from noti where userId = ?", userId)
	if err != nil {
		return nil, err
	}

	notis := []types.NotiPayload{}
	for rows.Next() {
		n, err := scanRowIntoNoti(rows)
		if err != nil {
			return nil, err
		}
		notis = append(notis, *n)
	}

	return notis, nil
}

func scanRowIntoNoti(rows *sql.Rows) (*types.NotiPayload, error) {
	noti := new(types.NotiPayload)
	err := rows.Scan(
		&noti.ID,
		&noti.UserID,
		&noti.Ip,
		&noti.Message,
		&noti.CreatedAt,
	)
	return noti, err
}
