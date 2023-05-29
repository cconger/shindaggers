package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cconger/shindaggers/pkg/db"
	"github.com/cconger/shindaggers/pkg/twitch"

	"github.com/gorilla/mux"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Webserver for fronting the database.
// Basic unauthed web_paths and a webhook to create a new pull

func main() {
	devMode := flag.Bool("dev", false, "enable dev mode which reloads the templates at runtime to allow rapid iteration")
	isolated := flag.Bool("nodb", false, "enable the application to use mock intefaces to dependencies, allows you to develop without having access to other services")
	flag.Parse()

	if *devMode {
		log.Println("Developer mode enabled! templates reloaded on every request")
	}

	clientID := os.Getenv("TWITCH_CLIENT_ID")
	clientSecret := os.Getenv("TWITCH_SECRET")
	webhookSecret := os.Getenv("WEBHOOK_SECRET")
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	var blobClient blobClient
	var twitchClient twitchClient
	var dbClient db.KnifeDB
	var err error

	if *isolated {
		blobClient = &mockBlobClient{}
		twitchClient = &twitch.MockClient{}
		dbClient = &db.MockDB{}
	} else {
		// Credentials to be able to upload images
		r2AccessKey := os.Getenv("CLOUDFLARE_SECRET")
		r2KeyID := os.Getenv("CLOUDFLARE_CLIENT_ID")
		storageEndpoint := os.Getenv("STORAGE_ENDPOINT")

		blobClient, err = minio.New(storageEndpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(r2KeyID, r2AccessKey, ""),
			Secure: true,
		})
		if err != nil {
			log.Printf("Could not initialize minioClient, image uploading will not work: %s", err)
			blobClient = nil
		}

		twitchClient, err = twitch.NewClient(clientID, clientSecret)
		if err != nil {
			log.Fatalf("failed to create twitchclient: %s", err)
		}

		dbClient, err = db.NewSDDB(os.Getenv("DSN"))
		if err != nil {
			log.Fatal(err)
		}
	}

	s := Server{
		devMode:        *devMode,
		db:             dbClient,
		webhookSecret:  webhookSecret,
		twitchClientID: clientID,
		twitchClient:   twitchClient,
		minioClient:    blobClient,
		bucketName:     "sd-images",

		baseURL: baseURL,
	}

	r := mux.NewRouter()
	r.HandleFunc("/", s.IndexHandler)
	r.HandleFunc("/me", s.Me)
	r.HandleFunc("/user/{id}", s.UserHandler)
	r.HandleFunc("/knife/{id:[0-9]+}", s.KnifeHandler)

	r.HandleFunc("/catalog/{id:[0-9]+}", s.CatalogView)
	r.HandleFunc("/catalog", s.CatalogHandler)

	r.HandleFunc("/oauth/redirect", s.OAuthHandler)
	r.HandleFunc("/pull/{token}", s.PullHandler).Methods(http.MethodPost)

	r.HandleFunc("/admin", s.OnlyAdmin(s.AdminIndex))
	r.HandleFunc("/admin/knife", s.OnlyAdmin(s.AdminKnifeList)).Methods(http.MethodGet)
	r.HandleFunc("/admin/knife", s.OnlyAdmin(s.AdminCreateKnife)).Methods(http.MethodPost)
	r.HandleFunc("/admin/knife/{id:[0-9]+}", s.OnlyAdmin(s.AdminKnife)).Methods(http.MethodGet)
	r.HandleFunc("/admin/knife/{id:[0-9]+}", s.OnlyAdmin(s.AdminUpdateKnife)).Methods(http.MethodPut)
	r.HandleFunc("/admin/knife/{id:[0-9]+}", s.OnlyAdmin(s.AdminDeleteKnife)).Methods(http.MethodDelete)

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
	log.Println("Interrupt signal recieved. Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	dbClient.Close(ctx)
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Error stopping http server: %v", err)
	}
}
