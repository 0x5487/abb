package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jasonsoft/abb/abb"
	"github.com/jasonsoft/abb/config"
	"github.com/jasonsoft/abb/identity"
	"github.com/jasonsoft/log"
	"github.com/jasonsoft/napnap"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			// unknown error
			err, ok := r.(error)
			if !ok {
				err = fmt.Errorf("unknown error: %v", err)
			}
			log.Errorf("unknown error: %v", err)
		}
	}()

	config := config.Config()

	// set up the log
	log.SetAppID("abb") // unique id for the app

	// set up the napnap
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGTERM)
	nap := napnap.New()
	nap.Use(napnap.NewHealth())

	corsOpts := napnap.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders: []string{"*"},
	}
	nap.Use(napnap.NewCors(corsOpts))
	nap.Use(abb.NewErrorHandlingMiddleware())
	nap.Use(identity.NewPublicIdentityRouter())

	// private router which needs to be authorized.
	jwtOpts := identity.JwtOptions{
		SecretKey:     config.Jwt.SecretKey,
		DurationInMin: config.Jwt.DurationInMin,
	}
	nap.Use(identity.NewJWTMiddleware(jwtOpts))
	nap.Use(identity.NewPrivateIdentityRouter())
	nap.Use(abb.NewAbbRouter())

	httpEngine := napnap.NewHttpEngine(":10214")
	log.Info("abb api starting")
	go func() {
		// service connections
		err := nap.Run(httpEngine)
		if err != nil {
			log.Error(err)
		}
	}()

	<-stopChan
	log.Info("Shutting down server...")

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	httpEngine.Shutdown(ctx)

	log.Info("gracefully stopped")
}
