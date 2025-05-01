package schedule

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

func (s *Store) CreateSchedule(payload types.Schedule) error {
	_, err := s.db.Exec("INSERT INTO schedules (deviceId, userId, action, scheduledTime, repeatDays) VALUES (?,?,?,?,?)", payload.DeviceID, payload.UserID, payload.Action, payload.ScheduledTime, payload.RepeatDays)
	return err
}

func (s *Store) GetAllActiveSchedule() ([]types.Schedule, error) {
	query := `
	SELECT id, deviceId, userId, action, scheduledTime, repeatDays, timezone
	FROM schedules
	WHERE isActive = TRUE;
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}

	schedules := []types.Schedule{}

	for rows.Next() {
		s, err := scanRowIntoSchedule(rows)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, *s)
	}

	return schedules, nil
}

func (s *Store) GetScheduleByFeedId(feed_id string) ([]types.Schedule, error) {
	query := `
		SELECT id, deviceId, userId, action, scheduledTime, repeatDays, timezone
		FROM schedules
		WHERE deviceId = ?
	`

	rows, err := s.db.Query(query, feed_id)
	if err != nil {
		return nil, err
	}

	schedules := []types.Schedule{}
	for rows.Next() {
		s, err := scanRowIntoSchedule(rows)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, *s)
	}

	return schedules, nil
}

func (s *Store) GetScheduleByID(id int) (types.Schedule, error) {
	var sch types.Schedule
	query := `
		SELECT id, deviceId, userId, action, scheduledTime, repeatDays, timezone, isActive
		FROM schedules WHERE id = ?
	`
	err := s.db.QueryRow(query).Scan(
		&sch.ID,
		&sch.DeviceID,
		&sch.UserID,
		&sch.Action,
		&sch.ScheduledTime,
		&sch.RepeatDays,
		&sch.Timezone,
		&sch.IsActive,
	)
	return sch, err
}

func (s *Store) UpdateSchedule(payload types.Schedule) error {
	query := `
	UPDATE schedules
	SET deviceId = ?, 
		userId = ?, 
		action = ?, 
		scheduledTime = ?, 
		repeatDays = ?, 
		timezone = ?, 
		isActive = ?
	WHERE id = ?;
	`

	_, err := s.db.Exec(query,
		payload.DeviceID,
		payload.UserID,
		payload.Action,
		payload.ScheduledTime,
		payload.RepeatDays,
		payload.Timezone,
		payload.IsActive,
		payload.ID,
	)

	return err
}


func (s *Store) RemoveSchedule(id int) error {
	_, err := s.db.Exec("DELETE FROM schedules WHERE id = ?", id)
	return err
}



func scanRowIntoSchedule(rows *sql.Rows) (*types.Schedule, error) {
	s := new(types.Schedule)
	err := rows.Scan(
		&s.ID,
		&s.DeviceID,
		&s.UserID,
		&s.Action,
		&s.ScheduledTime,
		&s.RepeatDays,
		&s.Timezone,
	)
	if err != nil {
		fmt.Println("Scan error:", err)
		return nil, err
	}

	return s, nil
}
