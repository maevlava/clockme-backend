package server

import (
	"github.com/clockme/clockme-backend/internal/config"
	"github.com/clockme/clockme-backend/internal/db"
	"github.com/clockme/clockme-backend/internal/middleware"
	"github.com/clockme/clockme-backend/internal/users"
	_ "github.com/lib/pq"
	"net/http"
)

type ClockmeServer struct {
	Address     string
	router      *http.ServeMux
	userHandler *users.UserHandler
	cfg         *config.Config
	db          *db.Queries
}

func NewClockmeServer(cfg *config.Config, db *db.Queries) *ClockmeServer {
	userHandler := users.NewUserHandler(db)

	c := &ClockmeServer{
		Address:     ":" + cfg.BackendPort,
		router:      http.NewServeMux(),
		cfg:         cfg,
		db:          db,
		userHandler: userHandler,
	}

	c.routes()
	return c
}
func (c *ClockmeServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.router.ServeHTTP(w, r)
}
func (c *ClockmeServer) routes() {

	// Users
	createUserHandler := http.HandlerFunc(c.userHandler.CreateUser)
	c.router.Handle("POST /api/v1/users", middleware.EnableCORS(createUserHandler))

	getAllUsersHandler := http.HandlerFunc(c.userHandler.GetUsers)
	c.router.Handle("GET /api/v1/users", middleware.EnableCORS(getAllUsersHandler))

	getUserHandler := http.HandlerFunc(c.userHandler.GetUser)
	c.router.Handle("GET /api/v1/users/{userID}", middleware.EnableCORS(getUserHandler))

	updateUserHandler := http.HandlerFunc(c.userHandler.UpdateUser)
	c.router.Handle("PUT /api/v1/users/{userID}", middleware.EnableCORS(updateUserHandler))

	deleteUserHandler := http.HandlerFunc(c.userHandler.DeleteUser)
	c.router.Handle("DELETE /api/v1/users/{userID}", middleware.EnableCORS(deleteUserHandler))
}
