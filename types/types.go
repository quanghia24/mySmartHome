package types

import "time"

type UserStore interface {
	GetUserByEmail(email string) (*User, error)
	GetUserByID(id int) (*User, error)
	CreateUser(User) error
}

type RoomStore interface {
	CreateRoom(Room) error
	GetRoomsByUserID(userId int) ([]RoomInfoPayload, error)
}

type DeviceStore interface {
	CreateDevice(Device) error
	GetDevicesByUserID(userId int) ([]DeviceDataPayload, error)
	GetDevicesByFeedID(feedId int) (*DeviceDataPayload, error)
	GetDevicesInRoomID(id int) ([]DeviceDataPayload, error)
}

type LogStore interface {
	CreateLog(Log) error
	GetLogsByID(id int) ([]Log, error)
}

type ProductStore interface {
	GetProducts() ([]Product, error)
	GetProductsByIDs(ps []int) ([]Product, error)
	CreateProduct(CreateProductPayload) error
	UpdateProduct(Product) error
}

type OrderStore interface {
	CreateOrder(Order) (int, error)
	CreateOrderItem(OrderItem) error
}

type Log struct {
	ID        int       `json:"id"`
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	DeviceID  int       `json:"deviceID"`
	UserID    int       `json:"userID"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"createdAt"`
}

type Room struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	UserID int    `json:"userID"`
}

type Device struct {
	FeedId  int    `json:"feedId"`
	FeedKey string `json:"feedKey"`
	Title   string `json:"title"`
	Type    string `json:"type"`
	UserID  int    `json:"userID"`
	RoomID  int    `json:"roomID"`
}

type Order struct {
	ID        int       `json:"id"`
	UserID    int       `json:"userID"`
	Total     int       `json:"total"`
	Status    string    `json:"status"`
	Address   string    `json:"address"`
	CreatedAt time.Time `json:"createdAt"`
}

type OrderItem struct {
	ID        int       `json:"id"`
	OrderID   int       `json:"orderID"`
	ProductID int       `json:"productID"`
	Quantity  int       `json:"quantity"`
	Price     float64   `json:"price"`
	CreatedAt time.Time `json:"createdAt"`
}

type CreateProductPayload struct {
	Name        string  `json:"name" validate:"required"`
	Description string  `json:"description" validate:"required"`
	Image       string  `json:"image" validate:"required"`
	Price       float64 `json:"price" validate:"required"`
	Quantity    int     `json:"quantity" validate:"required"`
}

type Product struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Image       string    `json:"image"`
	Price       float64   `json:"price"`
	Quantity    int       `json:"quantity"`
	CreatedAt   time.Time `json:"createdAt"`
}

type User struct {
	ID        int       `json:"id"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"createdAt"`
}

type RegisterUserPayload struct {
	FirstName string `json:"firstName" validate:"required"`
	LastName  string `json:"lastName" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=3,max=90"`
}

type LoginUserPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type CartItem struct {
	ProductID int `json:"productID"`
	Quantity  int `json:"quantity"`
}
type CartCheckoutPayload struct {
	Items []CartItem `json:"items" validate:"required"`
}

type CreateRoomPayload struct {
	Title string `json:"title" validate:"required"`
}

type CreateDevicePayload struct {
	FeedID  int    `json:"feedId" validate:"required"`
	FeedKey string `json:"feedkey" validate:"required"`
	Title   string `json:"title" validate:"required"`
	Type    string `json:"type" validate:"required"`
	RoomID  int    `json:"roomID" validate:"required"`
}

type DeviceDataPayload struct {
	FeedID    string    `json:"feedId"`
	Value     string    `json:"value" validate:"required"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
}

type LogCreatePayload struct {
	Type     string `json:"type"`
	Message  string `json:"message"`
	DeviceID int    `json:"deviceId"`
	UserID   int    `json:"userId"`
	Value    string `json:"value"`
}

type RoomInfoPayload struct {
	ID    int    `json:"id"`
	Title string `json:"title"`

	FanCount    int  `json:"fanCount"`
	LightCount  int  `json:"lightCount"`
	DoorCount   int  `json:"doorCount"`
	SensorCount int `json:"sensorCount"`
	FanStatus   int `json:"fanStatus"`
	LightStatus int `json:"lightStatus"`
	DoorStatus  int `json:"doorStatus"`
	SensorStatus int `json:"sensorStatus"`
}
