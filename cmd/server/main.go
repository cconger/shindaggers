package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/cconger/shindaggers/pkg/db"
	"github.com/cconger/shindaggers/pkg/twitch"

	"github.com/bwmarrin/snowflake"
	"github.com/gorilla/mux"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type blobClient interface {
	PutObject(context.Context, string, string, io.Reader, int64, minio.PutObjectOptions) (minio.UploadInfo, error)
}

type mockBlobClient struct{}

func (m *mockBlobClient) PutObject(ctx context.Context, bucket string, file string, contents io.Reader, size int64, options minio.PutObjectOptions) (minio.UploadInfo, error) {
	return minio.UploadInfo{}, fmt.Errorf("cannot upload files in dev mode")
}

type UserID struct {
	TwitchID   string
	InternalID int64
	Name       string
}

func (id *UserID) IsTwitch() bool {
	return id.TwitchID != ""
}

func (id *UserID) IsInternal() bool {
	return id.InternalID != 0
}

func (id *UserID) IsName() bool {
	return id.Name != ""
}

func createAuthToken() ([]byte, error) {
	// Secure random => sha256
	entropy := make([]byte, 100)
	_, err := rand.Read(entropy)
	if err != nil {
		return nil, err
	}
	token := sha256.Sum256(entropy)
	return token[:], nil
}

func ParseUserID(str string) UserID {
	if strings.HasPrefix(str, "twitch:") {
		return UserID{
			TwitchID: str[7:],
		}
	}

	if regexp.MustCompile(`^\d+$`).MatchString(str) {
		n, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			log.Printf("Unable to parse a numeric id? what the hell")
			return UserID{}
		}
		return UserID{
			InternalID: n,
		}
	}

	return UserID{
		Name: str,
	}
}

func (id *UserID) String() string {
	if id.IsTwitch() {
		return "twitch:" + id.TwitchID
	}

	if id.IsInternal() {
		return strconv.FormatInt(id.InternalID, 10)
	}

	if id.IsName() {
		return id.Name
	}

	return "UNKNOWN"
}

// Webserver for fronting the database.
// Basic unauthed web_paths and a webhook to create a new pull

func main() {
	devMode := flag.Bool("dev", false, "enable dev mode which reloads the templates at runtime to allow rapid iteration")
	isolated := flag.Bool("nodb", false, "enable the application to use mock intefaces to dependencies, allows you to develop without having access to other services")
	flag.Parse()

	discordWebhook := os.Getenv("DISCORD_WEBHOOK")
	clientID := os.Getenv("TWITCH_CLIENT_ID")
	clientSecret := os.Getenv("TWITCH_SECRET")
	webhookSecret := os.Getenv("WEBHOOK_SECRET")
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	var blobClient blobClient
	var twitchClient twitch.TwitchClient
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

	alloc_id := os.Getenv("FLY_ALLOC_ID")
	if alloc_id == "" {
		alloc_id = "3ff" // Dev Node
	}
	node_value, err := strconv.ParseInt(alloc_id[len(alloc_id)-3:], 16, 64)
	if err != nil {
		log.Fatalf("Unable to parse the node_value from %s", alloc_id)
	}
	log.Println("Running with node id:", node_value)

	node, err := snowflake.NewNode(node_value % 1024)
	if err != nil {
		log.Fatal("Unable to create node generator", err)
	}

	s := Server{
		devMode:        *devMode,
		db:             dbClient,
		webhookSecret:  webhookSecret,
		twitchClientID: clientID,
		twitchClient:   twitchClient,
		minioClient:    blobClient,
		bucketName:     "sd-images",
		idGenerator:    node,
		discordWebhook: discordWebhook,

		baseURL: baseURL,
	}

	r := mux.NewRouter()

	r.HandleFunc("/oauth/login", s.LoginHandler).Methods(http.MethodGet)
	r.HandleFunc("/oauth/handler", s.LoginResponseHandler).Methods(http.MethodGet)

	r.HandleFunc("/api/catalog", s.getCollection).Methods(http.MethodGet)
	r.HandleFunc("/api/collectable/{id:[0-9]+}", s.getCollectable).Methods(http.MethodGet)
	r.HandleFunc("/api/collectable", s.createCollectable).Methods(http.MethodPost)
	r.HandleFunc("/api/issued/{id:[0-9]+}", s.getIssuedCollectable).Methods(http.MethodGet)

	r.HandleFunc("/api/latest", s.getLatest).Methods(http.MethodGet)
	r.HandleFunc("/api/user/me", s.getLoggedInUser).Methods(http.MethodGet)

	r.HandleFunc("/api/user/{userid}", s.getUser).Methods(http.MethodGet)
	r.HandleFunc("/api/user/{userid}/equipped", s.getEquippedForUser).Methods(http.MethodGet)
	r.HandleFunc("/api/user/{userid}/collection", s.getUserCollection).Methods(http.MethodGet)
	r.HandleFunc("/api/user/{userid}/stats", s.getUserStats).Methods(http.MethodGet)

	// Search Users
	r.HandleFunc("/api/users", s.getUsers).Methods(http.MethodGet)

	// Legacy URL to be deleted
	r.HandleFunc("/pull/{token}", s.PullHandler).Methods(http.MethodPost)

	r.HandleFunc("/api/pull/{token}", s.PullHandler).Methods(http.MethodPost)
	r.HandleFunc("/api/randompull/{token}", s.RandomPullHandler).Methods(http.MethodPost)
	r.HandleFunc("/api/user/equip", s.EquipHandler).Methods(http.MethodPost)

	// ADMIN APIs
	r.HandleFunc("/api/admin/collectables", s.adminListCollectables).Methods(http.MethodGet)
	// Create Collectable
	r.HandleFunc("/api/admin/collectable", s.adminCreateCollectable).Methods(http.MethodPost)
	// Get Collectable
	r.HandleFunc("/api/admin/collectable/{id}", s.adminGetCollectable).Methods(http.MethodGet)
	// Modify Collectable
	r.HandleFunc("/api/admin/collectable/{id}", s.adminUpdateCollectable).Methods(http.MethodPut)
	// Modify Collectable
	r.HandleFunc("/api/admin/collectable/{id}/approve", s.adminApproveCollectable).Methods(http.MethodPost)
	// Delete Collectable
	r.HandleFunc("/api/admin/collectable/{id}", s.adminDeleteCollectable).Methods(http.MethodDelete)

	// Issue IssuedCollectable to User
	r.HandleFunc("/api/admin/issue", s.adminIssueCollectable).Methods(http.MethodPost)

	// Revoke IssuedCollectable
	r.HandleFunc("/api/admin/issued/{id}", s.adminRevokeIssuedCollectable).Methods(http.MethodDelete)

	// IssueConfig changes manages the weights of random pulls
	r.HandleFunc("/api/admin/issueconfig", s.adminGetIssueConfig).Methods(http.MethodGet)
	// ChangeIssueConfig
	r.HandleFunc("/api/admin/issueconfig", s.adminUpdateIssueConfig).Methods(http.MethodPut)

	// Image Upload
	r.HandleFunc("/api/image", s.ImageUpload).Methods(http.MethodPost)
	r.HandleFunc("/api/combat/report", s.CombatReportHandler).Methods(http.MethodPost)

	// Overlay is a mini SPA for OBS
	r.HandleFunc("/overlay/{id}", s.overlayHandler).Methods(http.MethodGet)
	r.PathPrefix("/assets").HandlerFunc(s.assetHandler)
	r.PathPrefix("/").HandlerFunc(s.spaHandler)

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
