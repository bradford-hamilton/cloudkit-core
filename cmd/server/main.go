package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bradford-hamilton/cloudkit-core/internal/cloudkit"
	"github.com/bradford-hamilton/cloudkit-core/internal/server"
	"github.com/bradford-hamilton/cloudkit-core/internal/storage"
	"github.com/sirupsen/logrus"
)

const tmpConnStr = "206.189.218.106:16509"

func main() {
	// TODO: env var check
	// TODO: switch gin to release (prod) mode in prod
	// TODO: hooks, config, etc for logging

	log := logrus.New()
	log.WithFields(nil).Info("Application initializing...")

	db, err := storage.NewDatabase()
	if err != nil {
		log.Panicf("failed to initialize PostgreSQL connection", err)
	}

	ckm, err := cloudkit.NewVMManager(tmpConnStr, log)
	if err != nil {
		log.Panicf("failed to initialize new cloudkit", err)
	}

	app := server.New(ckm, db, log)
	httpSrv := &http.Server{Addr: ":4000", Handler: app.Router()}

	// Initialize server in a goroutine so we don't block the graceful shutdown handling below.
	go func() {
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpSrv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
}
