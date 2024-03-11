package main

import (
	"encoding/base64"
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
	model "github.com/cconger/shindaggers/pkg/db/.gen/postgres/public/model"
	"github.com/cconger/shindaggers/pkg/twitch"

	"github.com/bwmarrin/snowflake"
	"github.com/minio/minio-go/v7"
)

type Server struct {
	devMode        bool
	db             db.PostgresDB
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
	user, err := s.db.GetUser(ctx, db.GetUserOptions{
		TwitchID: twitchUser.ID,
	})
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			user, err = s.db.CreateUser(ctx, db.User{
				Users: model.Users{
					ID:       s.idGenerator.Generate().Int64(),
					Name:     twitchUser.DisplayName,
					TwitchID: &twitchUser.ID,
				},
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

	if user.Name != twitchUser.DisplayName {
		// We need to update this user
		user.Name = twitchUser.DisplayName

		user, err = s.db.UpdateUser(ctx, *user)
		if err != nil {
			slog.Error("failed updating usernames for user", "id", user.ID, "err", err)
		}
	}

	token, err := createAuthToken()
	if err != nil {
		slog.Error("creating auth token", "err", err)
		http.Redirect(w, r, s.baseURL, http.StatusFound)
		return
	}

	// Store access token and refresh token
	err = s.db.SaveAuth(
		ctx,
		db.UserAuth{
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
