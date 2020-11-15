package server

import (
	"os"

	"github.com/bradford-hamilton/cloudkit-core/internal/cloudkit"
	"github.com/bradford-hamilton/cloudkit-core/internal/storage"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	ginlogrus "github.com/toorop/gin-logrus"
)

// App descirbes our main struct which holds all of the important dependencies and is
// used to handle requests and execute actions.
type App struct {
	router  *gin.Engine
	manager cloudkit.VMController
	storage storage.Datastore
	logger  *logrus.Logger
	baseURL string
}

// New spins up a new gin router, initializes all the application routes, and returns
// a new App struct with the gin router attached.
func New(ckm cloudkit.VMController, db storage.Datastore, log *logrus.Logger) *App {
	r := gin.New()
	r.Use(gin.Recovery(), cors.Default(), ginlogrus.Logger(log))

	app := App{
		router:  r,
		storage: db,
		manager: ckm,
		logger:  log,
		baseURL: os.Getenv("CLOUDKIT_BASE_URL"),
	}
	app.initializeRoutes()

	return &app
}

func (a *App) initializeRoutes() {
	a.router.GET("/ping", a.ping)

	v1 := a.router.Group("/api/v1")
	{
		v1.GET("/vms", a.getVMs)
		v1.POST("/vms", a.createVM)
		v1.GET("/vms/:domain_id", a.getVMByDomainID)
	}
}

// Router returns access to the router (*gin.Engine) field.
func (a *App) Router() *gin.Engine {
	return a.router
}
