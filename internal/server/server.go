package server

import (
	"os"

	"github.com/bradford-hamilton/cloudkit-core/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	ginlogrus "github.com/toorop/gin-logrus"
)

// API ...
type API struct {
	router  *gin.Engine
	db      storage.Datastore
	logger  *logrus.Logger
	baseURL string
}

// New ...
func New(db storage.Datastore, log *logrus.Logger) (*API, error) {
	r := gin.New()
	r.Use(ginlogrus.Logger(log), gin.Recovery())

	api := API{
		router:  r,
		logger:  log,
		baseURL: os.Getenv("CLOUDKIT_BASE_URL"),
		db:      db,
	}
	api.initializeRoutes()

	return &api, nil
}

func (a *API) initializeRoutes() {
	a.router.GET("/ping", a.ping)
}

// Router returns access to the router (*gin.Engine) field
func (a *API) Router() *gin.Engine {
	return a.router
}
