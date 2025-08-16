package server

import (
	"github.com/clockme/clockme-backend/internal/features/projects"
	"github.com/clockme/clockme-backend/internal/features/tasks"
	"github.com/clockme/clockme-backend/internal/features/users"
	"github.com/clockme/clockme-backend/internal/shared/config"
	"github.com/clockme/clockme-backend/internal/shared/db"
	"github.com/clockme/clockme-backend/internal/shared/middleware"
	_ "github.com/lib/pq"
	"net/http"
)

type ClockmeServer struct {
	Address        string
	router         *http.ServeMux
	userHandler    *users.UserHandler
	projectHandler *projects.ProjectHandler
	taskHandler    *tasks.TaskHandler
	cfg            *config.Config
	db             *db.Queries
}

func NewClockmeServer(cfg *config.Config, db *db.Queries) *ClockmeServer {
	userHandler := users.NewUserHandler(db)
	projectHandler := projects.NewProjectHandler(db)
	taskHandler := tasks.NewTaskHandler(db)

	c := &ClockmeServer{
		Address:        ":" + cfg.BackendPort,
		router:         http.NewServeMux(),
		cfg:            cfg,
		db:             db,
		userHandler:    userHandler,
		projectHandler: projectHandler,
		taskHandler:    taskHandler,
	}

	c.routes()
	return c
}
func (c *ClockmeServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.router.ServeHTTP(w, r)
}
func (c *ClockmeServer) routes() {
	mw := func(next http.Handler) http.Handler {
		return middleware.EnableCORS(next)
	}
	c.userHandler.RegisterRoutes(c.router, mw)
	c.projectHandler.RegisterRoutes(c.router, mw)
	c.taskHandler.RegisterRoutes(c.router, mw)
}
