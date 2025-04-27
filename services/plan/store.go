package plan

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

func (s *Store) CreatePlan(plan types.Plan) error {
	_, err := s.db.Exec("INSERT INTO plans (sensorId, lower, upper) VALUES (?,?,?)", plan.SensorID, plan.Lower, plan.Upper)
	return err
}

func (s *Store) RemovePlan(sensorID int) error {
	_, err := s.db.Exec("DELETE FROM plans WHERE sensorId = ?", sensorID)
	return err
}

func (s *Store) GetPlansByFeedID(sensorID int) (*types.Plan, error) {
	row := s.db.QueryRow(`
        SELECT id, sensorId, lower, upper, createdAt
        FROM plans
        WHERE sensorId = ?
        ORDER BY createdAt DESC
        LIMIT 1
    `, sensorID)

	var p types.Plan
    err := row.Scan(&p.ID, &p.SensorID, &p.Lower, &p.Upper, &p.CreatedAt)
    if err != nil {
        return nil, err
    }
    return &p, nil
}

// func scanIntoPlan(rows *sql.Rows) (*types.Plan, error) {
// 	p := new(types.Plan)

// 	err := rows.Scan(
// 		&p.ID,
// 		&p.SensorID,
// 		&p.Lower,
// 		&p.Upper,
// 		p.CreatedAt,
// 	)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return p, nil
// }
