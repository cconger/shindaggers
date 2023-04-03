package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cconger/shindaggers/pkg/db"

	"github.com/gorilla/mux"
)

// Webserver for fronting the database.
// Basic unauthed web_paths and a webhook to create a new pull

func main() {
	devMode := flag.Bool("dev", false, "enable dev mode which reloads the templates at runtime to allow rapid iteration")
	flag.Parse()

	if *devMode {
		log.Println("Developer mode enabled!")
	}

	db, err := db.NewSDDB(os.Getenv("DSN"))
	if err != nil {
		log.Fatal(err)
	}
	s := Server{
		devMode:       *devMode,
		db:            db,
		webhookSecret: os.Getenv("WEBHOOK_SECRET"),
	}

	r := mux.NewRouter()
	r.HandleFunc("/", s.IndexHandler)
	r.HandleFunc("/user/{id}", s.UserHandler)
	r.HandleFunc("/knife/{id:[0-9]+}", s.KnifeHandler)
	r.HandleFunc("/pull/{token}", s.PullHandler).Methods(http.MethodPost)

	http.Handle("/", r)

	interrupt := make(chan os.Signal, 1)

	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	srv := &http.Server{
		Addr: ":8080",
	}

	go func() {
		log.Println("starting webserver")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Error running server: %s", err)
		}
	}()

	<-interrupt
	fmt.Println("Interrupt signal recieved. Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	db.Close(ctx)
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Error stopping http server: %v", err)
	}
}
