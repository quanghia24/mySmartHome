package api

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/quanghia24/mySmartHome/services/cart"
	"github.com/quanghia24/mySmartHome/services/device"
	"github.com/quanghia24/mySmartHome/services/log_device"
	"github.com/quanghia24/mySmartHome/services/order"
	"github.com/quanghia24/mySmartHome/services/product"
	"github.com/quanghia24/mySmartHome/services/room"
	"github.com/quanghia24/mySmartHome/services/user"
	"github.com/rs/cors"
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

	subrouter := router.PathPrefix("/api/v1").Subrouter()

	userStore := user.NewStore(s.db)
	userHanlder := user.NewHandler(userStore)
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

	logStore := log_device.NewStore(s.db)
	logHandler := log_device.NewHandler(logStore, userStore)
	logHandler.RegisterRoutes(subrouter)

	deviceStore := device.NewStore(s.db)
	deviceHandler := device.NewHandler(deviceStore, userStore, roomStore, logStore)
	deviceHandler.RegisterRoutes(subrouter)

	fmt.Println("Listening on port", s.addr)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodPut},
		AllowCredentials: true,
	})
	return http.ListenAndServe(s.addr, c.Handler(router))
}
