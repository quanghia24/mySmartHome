package api

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/quanghia24/mySmartHome/cmd/mqtt"
	"github.com/quanghia24/mySmartHome/services/cart"
	"github.com/quanghia24/mySmartHome/services/device"
	"github.com/quanghia24/mySmartHome/services/doorpwd"
	"github.com/quanghia24/mySmartHome/services/log_device"
	"github.com/quanghia24/mySmartHome/services/log_sensor"
	"github.com/quanghia24/mySmartHome/services/notification"
	"github.com/quanghia24/mySmartHome/services/order"
	"github.com/quanghia24/mySmartHome/services/plan"
	"github.com/quanghia24/mySmartHome/services/product"
	"github.com/quanghia24/mySmartHome/services/room"
	"github.com/quanghia24/mySmartHome/services/schedule"
	"github.com/quanghia24/mySmartHome/services/sensor"
	"github.com/quanghia24/mySmartHome/services/statistic"
	"github.com/quanghia24/mySmartHome/services/user"
)

type APIServer struct {
	addr string
	db   *sql.DB
}

func NewAPIServer(addr string, db *sql.DB) *APIServer {
	return &APIServer{
		addr: addr,
		db:   db,
	}
}

// initialize router
// register routes and their dependency -> make them services
func (s *APIServer) Run() error {
	router := mux.NewRouter()

	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Security-Policy", "default-src 'self' http://localhost:8000; script-src 'self' 'unsafe-inline' 'unsafe-eval'")
			next.ServeHTTP(w, r)
		})
	})

	mqttClient := mqtt.NewClient(s.db)

	subrouter := router.PathPrefix("/api/v1").Subrouter()

	notiStore := notification.NewStore(s.db)

	userStore := user.NewStore(s.db)
	userHanlder := user.NewHandler(userStore, notiStore)
	userHanlder.RegisterRoutes(subrouter)

	productStore := product.NewStore(s.db)
	productHandler := product.NewHandler(productStore)
	productHandler.RegisterRoutes(subrouter)

	orderStore := order.NewStore(s.db)

	cartHandler := cart.NewHandler(orderStore, productStore, userStore)
	cartHandler.RegisterRouter(subrouter)

	roomStore := room.NewStore(s.db)
	roomHandler := room.NewHandler(roomStore, userStore)
	roomHandler.RegisterRoutes(subrouter)

	logDeviceStore := log_device.NewStore(s.db)
	logDeviceHandler := log_device.NewHandler(logDeviceStore, userStore)
	logDeviceHandler.RegisterRoutes(subrouter)

	doorStore := doorpwd.NewStore(s.db)

	deviceStore := device.NewStore(s.db)
	deviceHandler := device.NewHandler(deviceStore, userStore, roomStore, logDeviceStore, doorStore, mqttClient)
	deviceHandler.RegisterRoutes(subrouter)

	logSensorStore := log_sensor.NewStore(s.db)
	logSensorHandler := log_sensor.NewHandler(logSensorStore)
	logSensorHandler.RegisterRoutes(subrouter)

	planStore := plan.NewStore(s.db)
	planHandler := plan.NewHandler(planStore)
	planHandler.RegisterRoutes(subrouter)

	sensorStore := sensor.NewStore(s.db)
	sensorHandler := sensor.NewHandler(sensorStore, userStore, logSensorStore, planStore, mqttClient)
	sensorHandler.RegisterRoutes(subrouter)

	go sensorHandler.StartSensorDataPolling()

	scheduleStore := schedule.NewStore(s.db)
	scheduleHandler := schedule.NewHandler(scheduleStore, deviceStore, logDeviceStore, doorStore, userStore)
	scheduleHandler.RegisterRoutes(subrouter)

	statisticHandler := statistic.NewHandler(logDeviceStore, logSensorStore, userStore, roomStore, deviceStore)
	statisticHandler.RegisterRoutes(subrouter)

	notiHandler := notification.NewHandler(notiStore, userStore)
	notiHandler.RegisterRoutes(subrouter)

	scheduleHandler.StartSchedule()

	// mqtt.ResubscribeDevices(deviceStore, mqttClient, logDeviceStore)
	// mqtt.ResubscribeSensors(sensorStore, deviceStore, mqttClient, planStore, logSensorStore, notiStore)
	// fmt.Println("Reconnected to mqtt")

	fmt.Println("Listening on port", s.addr)

	return http.ListenAndServe(s.addr,
		handlers.CORS(
			handlers.AllowedOrigins([]string{"*"}),
			handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
			handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}),
		)(router))
}
