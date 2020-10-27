package server

import (
	"os"

	"github.com/bradford-hamilton/cloudkit-core/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	ginlogrus "github.com/toorop/gin-logrus"
)

// New ...
func New(log *logrus.Logger, db storage.Datastore) (*API, error) {
	// Creates a router with no default middleware
	r := gin.New()

	r.Use(
		ginlogrus.Logger(log), // Integrate logging through logrus
		gin.Recovery(),        // recovers from any panics and writes a 500 if there was one.
	)

	baseURL := "http://localhost:4000"
	if os.Getenv("CLOUDKIT_CORE_ENVIRONMENT") == "production" {
		baseURL = "TODO"
	}

	a := API{
		Router:  r,
		logger:  log,
		baseURL: baseURL,
		db:      db,
	}
	a.initializeRoutes()

	return &a, nil
}

func (a *API) initializeRoutes() {
	a.Router.GET("/ping", a.ping)
}

// API ...
type API struct {
	Router  *gin.Engine
	db      storage.Datastore
	logger  *logrus.Logger
	baseURL string
}
