package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/cconger/shindaggers/pkg/db"
	"github.com/cconger/shindaggers/pkg/twitch"

	"github.com/bwmarrin/snowflake"
	"github.com/gorilla/mux"
	"github.com/minio/minio-go/v7"
)

type Server struct {
	devMode        bool
	db             db.KnifeDB
	webhookSecret  string
	twitchClientID string
	twitchClient   twitch.TwitchClient
	baseURL        string
	minioClient    blobClient
	bucketName     string
	idGenerator    *snowflake.Node
	discordWebhook string

	template *template.Template
}

func (s *Server) LoginHandler(w http.ResponseWriter, r *http.Request) {
	p := url.Values{
		"response_type": []string{"code"},
		"client_id":     []string{s.twitchClientID},
		"redirect_uri":  []string{fmt.Sprintf("%s/oauth/handler", s.baseURL)},
		"scope":         []string{""},
	}

	uri := "https://id.twitch.tv/oauth2/authorize?" + p.Encode()
	http.Redirect(
		w,
		r,
		uri,
		http.StatusFound,
	)
}

func (s *Server) LoginResponseHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	params := r.URL.Query()
	errored := params.Has("error")
	if errored {
		desc := params.Get("error_description")
		slog.Error("oAuth error", "err", desc)
		http.Redirect(w, r, s.baseURL, http.StatusFound)
		return
	}

	// Use code to get access token
	code := params.Get("code")
	if code == "" {
		slog.Error("code is empty")
		http.Redirect(w, r, s.baseURL, http.StatusFound)
		return
	}

	t, err := s.twitchClient.OAuthGetToken(
		ctx,
		code,
		fmt.Sprintf("%s/oauth/redirect", s.baseURL),
	)
	if err != nil {
		slog.Error("getting oauthtoken", "err", err)
		http.Redirect(w, r, s.baseURL, http.StatusFound)
		return
	}

	expiresAt := time.Now().Add(time.Duration(t.ExpiresIn) * time.Second)

	// Get user from twitch
	twitchClient := s.twitchClient.UserClient(&twitch.UserAuth{
		AccessToken:  t.AccessToken,
		RefreshToken: t.RefreshToken,
	})

	twitchUser, err := twitchClient.GetUser(ctx)
	if err != nil {
		slog.Error("getting twitch user", "err", err)
		http.Redirect(w, r, s.baseURL, http.StatusFound)
		return
	}

	// Get or create user in our db
	user, err := s.db.GetUserByTwitchID(ctx, twitchUser.ID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			user, err = s.db.CreateUser(ctx, &db.User{
				Name:       twitchUser.DisplayName,
				LookupName: twitchUser.Login,
				TwitchID:   twitchUser.ID,
			})
			if err != nil {
				slog.Error("creating user", "err", err)
				http.Redirect(w, r, s.baseURL, http.StatusFound)
				return
			}
		} else {
			slog.Error("getting user", "err", err)
			http.Redirect(w, r, s.baseURL, http.StatusFound)
			return
		}
	}

	token, err := createAuthToken()
	if err != nil {
		slog.Error("creating auth token", "err", err)
		http.Redirect(w, r, s.baseURL, http.StatusFound)
		return
	}

	// Store access token and refresh token
	_, err = s.db.SaveAuth(
		ctx,
		&db.UserAuth{
			UserID:       user.ID,
			Token:        token,
			AccessToken:  t.AccessToken,
			RefreshToken: t.RefreshToken,
			ExpiresAt:    expiresAt,
		},
	)
	if err != nil {
		slog.Error("saving auth token", "err", err)
		http.Redirect(w, r, s.baseURL, http.StatusFound)
		return
	}

	encodedToken := base64.URLEncoding.EncodeToString(token)

	baseURL := s.baseURL
	if s.devMode {
		baseURL = "http://localhost:3000"
	}

	http.Redirect(
		w,
		r,
		baseURL+"/login#token="+encodedToken,
		http.StatusFound,
	)
}

type PullRequest struct {
	TwitchID  string `json:"user_id"`
	Username  string `json:"username"`
	Knifename string `json:"knifename"`

	// These are ints to be more tolerant to the ingets model
	Verified   string `json:"verified"`
	Subscriber string `json:"sub_status"`

	Edition string `json:"edition"`
}

// PullHandler is the webhook handler for recording a knife pull after its been executed locally by the
// streamer
func (s *Server) PullHandler(w http.ResponseWriter, r *http.Request) {
	if s.webhookSecret == "" {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	ctx := r.Context()
	vars := mux.Vars(r)
	token := vars["token"]
	if token != s.webhookSecret {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var reqBody PullRequest
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		serveAPIErr(w, err, http.StatusBadRequest, "could not parse payload")
		return
	}
	defer r.Body.Close()

	slog.Info("PullHandler", "payload", reqBody)

	var user *db.User
	if reqBody.TwitchID != "" {
		user, err = s.db.GetUserByTwitchID(ctx, reqBody.TwitchID)
	} else {
		user, err = s.db.GetUserByUsername(ctx, reqBody.Username)
	}
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			slog.Info("PullHandler creating user", "twitchid", reqBody.TwitchID, "username", reqBody.Username)
			user, err = s.db.CreateUser(ctx, &db.User{
				Name:     reqBody.Username,
				TwitchID: reqBody.TwitchID,
			})
		}
		if err != nil {
			slog.Error("PullHandler getting user", "err", err)
			serveAPIErr(w, err, http.StatusBadRequest, "could not resolve user")
			return
		}
	}

	subscriber := reqBody.Subscriber == "1" || strings.ToLower(reqBody.Subscriber) == "true"
	verified := reqBody.Verified == "1" || strings.ToLower(reqBody.Verified) == "true"

	t, err := s.db.GetKnifeTypeByName(ctx, reqBody.Knifename)
	if err != nil {
		serveAPIErr(w, fmt.Errorf("unknown knife name %s: %w", reqBody.Knifename, err), http.StatusBadRequest, "Unknown knife name")
		return
	}

	k, err := s.db.IssueCollectable(ctx, &db.Knife{
		InstanceID: s.idGenerator.Generate().Int64(),
		ID:         t.ID,
		OwnerID:    user.ID,
		Subscriber: subscriber,
		Verified:   verified,
	}, "pull")
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			serveAPIErr(w, err, http.StatusBadRequest, "could not find knife or user")
			return
		}

		serveAPIErr(w, err, http.StatusInternalServerError, "unexpected error")
		return
	}

	err = json.NewEncoder(w).Encode(k)
	if err != nil {
		slog.Error("PullHandler serializing response", "err", err)
	}
}

func (s *Server) ImageUpload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, err := s.getAuthUser(ctx, r)
	if err != nil {
		serveAPIErr(w, err, http.StatusBadRequest, "could not determine user")
		return
	}

	err = r.ParseMultipartForm(32 << 20) // 32MB maximum file size
	if err != nil {
		serveAPIErr(w, err, http.StatusBadRequest, "could not parse form")
		return
	}

	// Get the file from the request
	file, handler, err := r.FormFile("image")
	if err != nil {
		serveAPIErr(w, err, http.StatusBadRequest, "no image detected")
		return
	}
	defer file.Close()

	// Check the file type and size
	if !strings.HasPrefix(handler.Header.Get("Content-Type"), "image/") {
		serveAPIErr(w, fmt.Errorf("file not image"), http.StatusBadRequest, "file not image")
		return
	}
	if handler.Size > 32<<20 {
		serveAPIErr(w, fmt.Errorf("file too large"), http.StatusBadRequest, "file too large")
		return
	}

	newImageID := s.idGenerator.Generate()
	uploadName := path.Base(handler.Filename)
	ext := path.Ext(handler.Filename)

	basename := newImageID.String() + ext
	if s.minioClient != nil {
		_, err = s.minioClient.PutObject(ctx, s.bucketName, path.Join("images", basename), file, handler.Size, minio.PutObjectOptions{
			ContentType: handler.Header.Get("Content-Type"),
		})
		if err != nil {
			serveAPIErr(w, err, http.StatusBadRequest, "error uploading image")
			return
		}
	}

	err = s.db.CreateImageUpload(ctx, newImageID.Int64(), user.ID, basename, uploadName)
	if err != nil {
		serveAPIErr(w, err, http.StatusInternalServerError, "error saving image upload image")
		return
	}

	// RETURN THE IMAGE TO PREVIEW
	serveAPIPayload(
		w,
		&struct {
			ImagePath string
			ImageURL  string
		}{
			ImagePath: basename,
			ImageURL:  "https://images.shindaggers.io/images/" + basename,
		},
	)
}
