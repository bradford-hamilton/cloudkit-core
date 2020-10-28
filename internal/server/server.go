package server

import (
	"os"

	"github.com/bradford-hamilton/cloudkit-core/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	ginlogrus "github.com/toorop/gin-logrus"
)

// App descirbes our main struct which holds all of the important dependencies and is
// used to handle requests and execute actions.
type App struct {
	router  *gin.Engine
	db      storage.Datastore
	logger  *logrus.Logger
	baseURL string
}

// New spins up a new gin router, initializes all the application routes, and returns
// a new App struct with the gin router attached.
func New(db storage.Datastore, log *logrus.Logger) *App {
	r := gin.New()
	r.Use(ginlogrus.Logger(log), gin.Recovery())

	api := App{
		router:  r,
		logger:  log,
		baseURL: os.Getenv("CLOUDKIT_BASE_URL"),
		db:      db,
	}
	api.initializeRoutes()

	return &api
}

func (a *App) initializeRoutes() {
	a.router.GET("/ping", a.ping)
}

// Router returns access to the router (*gin.Engine) field.
func (a *App) Router() *gin.Engine {
	return a.router
}
