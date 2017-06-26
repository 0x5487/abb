package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"gitlab.paradise-soft.com.tw/channels/cm/cm"

	"github.com/jasonsoft/abb"
	"github.com/jasonsoft/log"
	"github.com/jasonsoft/napnap"
)

func main() {
	// set up the log
	log.SetAppID("abb") // unique id for the app

	// set up the napnap
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, os.Kill)
	nap := napnap.New()
	nap.Use(napnap.NewHealth())
	nap.Use(abb.NewServiceRouter())

	httpEngine := napnap.NewHttpEngine(cm.Config().Vendor.Bind)
	log.Info("vendors starting")
	go func() {
		// service connections
		err := nap.Run(httpEngine)
		if err != nil {
			log.Error(err)
		}
	}()

	<-stopChan
	log.Info("Shutting down server...")

	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	httpEngine.Shutdown(ctx)

	log.Info("gracefully stopped")
}
