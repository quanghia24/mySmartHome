package types

import "time"

type UserStore interface {
	GetUserByEmail(email string) (*User, error)
	GetUserByID(id int) (*User, error)
	UpdateProfile(User) error
	CreateUser(User) error
}

type RoomStore interface {
	CreateRoom(Room) error
	GetRoomsByUserID(userId int) ([]RoomInfoPayload, error)
	GetDevicesByRoomId(roomId int) ([]int, error)
	UpdateRoom(Room) error
	DeleteRoom(roomId int, userId int) error
}

type DeviceStore interface {
	CreateDevice(Device) error
	GetAllDevices() ([]AllDeviceDataPayload, error)
	GetDevicesByUserID(userId int) ([]DeviceDataPayload, error)
	GetDevicesByFeedID(feedId int) (*DeviceDataPayload, error)
	GetDevicesInRoomID(id int) ([]DeviceDataPayload, error)
	GetDevicesByRoomIdAndType(roomId int, mtype string) ([]int, error)
	DeleteDevice(deviceId string, userId int) error
}

type SensorStore interface {
	CreateSensor(Sensor) error
	GetSensorByFeedID(feedId int) (*DeviceDataPayload, error)
	GetAllSensor() ([]Sensor, error)
}

type LogDeviceStore interface {
	CreateLog(LogDevice) error
	GetLogsByFeedID(feedId int) ([]LogDevice, error)
	GetLogsByFeedIDBetween(feedId int, start time.Time, end time.Time) ([]LogDevice, error)
	GetLogsByFeedID7Days(feedId int, end time.Time) ([]LogDevice, error)
	GetLogsByUserID(userId int) ([]LogDevice, error)
}

type LogSensorStore interface {
	CreateLogSensor(LogSensor) error
	GetLogSensorsByUserID(userId int) ([]LogSensor, error)
	GetSensorsByFeedIDBetween(feedId int, start time.Time, end time.Time) ([]LogSensor, error)
	GetLogSensorsLast7HoursByFeedID(feedId int, end time.Time) ([]LogSensor, error)
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

type DoorStore interface {
	CreatePassword(DoorPassword) error
	GetPassword(feedId int) (*DoorPassword, error)
}

type ScheduleStore interface {
	CreateSchedule(Schedule) error
	GetAllActiveSchedule() ([]Schedule, error)
	GetScheduleByFeedId(string) ([]Schedule, error)
	GetScheduleByID(id int) (Schedule, error)
	UpdateSchedule(Schedule) error
	RemoveSchedule(id int) error
}

type PlanStore interface {
	CreatePlan(Plan) error
	RemovePlan(int) error
	GetPlansByFeedID(int) (*Plan, error)
}

type NotiStore interface {
	CreateNotiIp(NotiIpPayload) error
	GetNotiIpByUserId(userId int) (*NotiIpPayload, error)

	CreateNoti(NotiPayload) error
	GetNotiByUserId(userId int) ([]NotiPayload, error)
}

type NotiPayload struct {
	ID        int       `json:"id"`
	UserID    int       `json:"userID"`
	Ip        string    `json:"ip"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type NotiIpPayload struct {
	UserID int    `json:"userID"`
	Ip     string `json:"ip"`
}

type RequestStatisticDevicePayload struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type Plan struct {
	ID        int       `json:"id"`
	SensorID  int       `json:"sensorId"`
	Lower     string    `json:"lower"`
	Upper     string    `json:"upper"`
	CreatedAt time.Time `json:"createdAt"`
}

type Schedule struct {
	ID            int    `json:"id"`
	DeviceID      int    `json:"deviceId"`
	UserID        int    `json:"userId"`
	Action        string `json:"action"`
	ScheduledTime string `json:"scheduledTime"` // stored as HH:MM:SS
	RepeatDays    string `json:"repeatDays"`    //e.g. Mon,Tue
	Timezone      string `json:"timezone"`
	IsActive      bool   `json:"isActive"`
}

type DoorPassword struct {
	ID        int       `json:"id"`
	FeedID    int       `json:"feedId"`
	PWD       string    `json:"pwd"`
	CreatedAt time.Time `json:"createdAt"`
}

type LogSensor struct {
	ID        int       `json:"id"`
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	SensorID  int       `json:"sensorID"`
	UserID    int       `json:"userID"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"createdAt"`
}

type LogDevice struct {
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
	Image  string `json:"image"`
}

type Device struct {
	FeedId  int    `json:"feedId"`
	FeedKey string `json:"feedKey"`
	Title   string `json:"title"`
	Type    string `json:"type"`
	UserID  int    `json:"userID"`
	RoomID  int    `json:"roomID"`
}

type Sensor struct {
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
	Avatar    string    `json:"avatar"`
	CreatedAt time.Time `json:"createdAt"`
}

type RegisterUserPayload struct {
	FirstName string `json:"firstName" validate:"required"`
	LastName  string `json:"lastName" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=3,max=90"`
	Ip        string `json:"ip,omitempty"`
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
	FeedID    int       `json:"feedId"`
	FeedKey   string    `json:"feedKey"`
	Value     string    `json:"value" validate:"required"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
}

type AllDeviceDataPayload struct {
	FeedID    int       `json:"feedId"`
	FeedKey   string    `json:"feedKey"`
	Value     string    `json:"value" validate:"required"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	UserID    int       `json:"userId"`
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

	FanCount    int `json:"fanCount"`
	FanStatus   int `json:"fanStatus"`
	LightCount  int `json:"lightCount"`
	LightStatus int `json:"lightStatus"`
	DoorCount   int `json:"doorCount"`
	DoorStatus  int `json:"doorStatus"`
	SensorCount int `json:"sensorCount"`
}

type SensorDataPayload struct {
	FeedId    int       `json:"feed_id"`
	FeedKey   string    `json:"feed_key"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
}

type UsageRequestPayload struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type TimeObject struct {
	Date  time.Time `json:"date"`
	Value string    `json:"value"`
}
