package order

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

func (s *Store) CreateOrder(order types.Order) (int, error) {
	res, err := s.db.Exec("INSERT INTO orders (userID, total, status, address) VALUES (?, ?, ?, ?)",
		order.UserID, order.Total, order.Status, order.Address)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), err
}

func (s *Store) CreateOrderItem(orderItem types.OrderItem) error {
	_, err := s.db.Exec("INSERT INTO orders_items (orderID, productID, quantity, price) VALUES (?, ?, ?, ?)",
		orderItem.OrderID, orderItem.ProductID, orderItem.Quantity, orderItem.Price)
	if err != nil {
		return err
	}
	return nil
}
