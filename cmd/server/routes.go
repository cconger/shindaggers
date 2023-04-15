package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/cconger/shindaggers/pkg/db"
	"github.com/cconger/shindaggers/pkg/twitch"

	"github.com/gorilla/mux"
	"github.com/minio/minio-go/v7"
)

const AUTH_COOKIE = "auth"

type KnifePage struct {
	db.Knife

	RarityClass string
}

type UserPage struct{}

func className(rarity string) string {
	switch rarity {
	case "Common":
		return "rarity-common"
	case "Uncommon":
		return "rarity-uncommon"
	case "Rare":
		return "rarity-rare"
	case "Super Rare":
		return "rarity-super-rare"
	case "Ultra Rare":
		return "rarity-ultra-rare"
	}
	return "rarity-common"
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

//go:embed templates/*
var templates embed.FS

func servererr(w http.ResponseWriter, err error, errorCode int) {
	w.WriteHeader(errorCode)
	fmt.Fprintf(w, "Error: %s", err)
}

type Server struct {
	devMode        bool
	db             db.KnifeDB
	webhookSecret  string
	twitchClientID string
	twitchClient   *twitch.Client
	baseURL        string
	minioClient    *minio.Client
	bucketName     string

	template *template.Template
}

func (s *Server) getTemplate() (*template.Template, error) {
	if s.devMode {
		return template.ParseGlob(path.Join("cmd", "server", "templates", "*"))
	}

	if s.template == nil {
		t, err := template.ParseFS(templates, path.Join("templates", "*"))
		if err != nil {
			return nil, err
		}
		s.template = t
	}

	return s.template, nil
}

func (s *Server) renderTemplate(w http.ResponseWriter, name string, payload interface{}) {
	t, err := s.getTemplate()
	if err != nil {
		servererr(w, err, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	err = t.ExecuteTemplate(w, name, payload)
	if err != nil {
		log.Printf("err executing template %s: %s", name, err)
	}
}

type PullListing struct {
	InstanceID  int
	Name        string
	Owner       string
	AbsTime     string
	TimeAgo     string
	ImageName   string
	RarityClass string
}

type IndexPayload struct {
	LoginURL string
	Pulls    []*PullListing
}

func (s *Server) loginWithTwitchURL() string {
	p := url.Values{
		"response_type": []string{"code"},
		"client_id":     []string{s.twitchClientID},
		"redirect_uri":  []string{fmt.Sprintf("%s/oauth/redirect", s.baseURL)},
		"scope":         []string{""},
	}

	return "https://id.twitch.tv/oauth2/authorize?" + p.Encode()
}

func (s *Server) IndexHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	knivesRes, err := s.db.GetLatestPulls(ctx)
	if err != nil {
		log.Printf("error getting latest pulls: %s", err)
	}

	pulls := make([]*PullListing, len(knivesRes))
	for i, k := range knivesRes {
		pulls[i] = &PullListing{
			InstanceID:  k.InstanceID,
			Name:        k.Name,
			Owner:       k.Owner,
			ImageName:   k.ImageName,
			RarityClass: className(k.Rarity),
			AbsTime:     k.ObtainedAt.String(),
			TimeAgo:     timeAgo(k.ObtainedAt),
		}
	}

	pl := IndexPayload{
		LoginURL: s.loginWithTwitchURL(),
		Pulls:    pulls,
	}

	s.renderTemplate(w, "index.html", pl)
}

func (s *Server) Me(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, err := s.getUserForRequest(ctx, r)
	if err != nil {
		servererr(w, fmt.Errorf("Unable to determine who you are: %w", err), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, s.baseURL+"/user/"+user.LookupName, http.StatusTemporaryRedirect)
}

func (s *Server) OAuthHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	params := r.URL.Query()
	errored := params.Has("error")
	if errored {
		desc := params.Get("error_description")
		log.Printf("OAUTH Error: %s", desc)
		http.Redirect(w, r, s.baseURL, http.StatusTemporaryRedirect)
		return
	}

	// Use code to get access token
	code := params.Get("code")
	if code == "" {
		log.Printf("code is empty")
		http.Redirect(w, r, s.baseURL, http.StatusTemporaryRedirect)
		return
	}

	t, err := s.twitchClient.OAuthGetToken(
		ctx,
		code,
		fmt.Sprintf("%s/oauth/redirect", s.baseURL),
	)
	if err != nil {
		log.Printf("error getting oauthtoken: %s", err)
		http.Redirect(w, r, s.baseURL, http.StatusTemporaryRedirect)
		return
	}

	expiresAt := time.Now().Add(time.Duration(t.ExpiresIn) * time.Second)

	// Get user from twitch
	twitchUser, err := s.twitchClient.GetUser(ctx, t.AccessToken)
	if err != nil {
		log.Printf("error getting twitch user %s", err)
		http.Redirect(w, r, s.baseURL, http.StatusTemporaryRedirect)
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
				log.Printf("error creating user %s", err)
				http.Redirect(w, r, s.baseURL, http.StatusTemporaryRedirect)
				return
			}
		} else {
			log.Printf("error getting user %s", err)
			http.Redirect(w, r, s.baseURL, http.StatusTemporaryRedirect)
			return
		}
	}

	token, err := createAuthToken()
	if err != nil {
		log.Printf("error creating auth token: %s", err)
		http.Redirect(w, r, s.baseURL, http.StatusTemporaryRedirect)
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
		log.Printf("error saving auth token: %s", err)
		http.Redirect(w, r, s.baseURL, http.StatusTemporaryRedirect)
		return
	}

	encodedToken := base64.RawStdEncoding.EncodeToString(token)

	// Set user cookie
	http.SetCookie(w, &http.Cookie{
		Name:     AUTH_COOKIE,
		Value:    encodedToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   !s.devMode,
	})

	// Redirect to user page
	http.Redirect(
		w,
		r,
		fmt.Sprintf("%s/user/%s", s.baseURL, user.LookupName),
		http.StatusTemporaryRedirect,
	)
}

type UserPagePayload struct {
	User   *db.User
	Knives []*KnifePage
}

func (s *Server) UserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)

	username := strings.ToLower(vars["id"])

	userRes, err := s.db.GetUserByUsername(ctx, username)
	if err != nil {
		servererr(w, err, http.StatusBadRequest)
		return
	}

	knivesRes, err := s.db.GetKnivesForUsername(ctx, username)
	if err != nil {
		servererr(w, err, http.StatusBadRequest)
		return
	}

	knives := make([]*KnifePage, len(knivesRes))
	for i, k := range knivesRes {
		knives[i] = &KnifePage{
			Knife:       *k,
			RarityClass: className(k.Rarity),
		}
	}

	payload := UserPagePayload{
		User:   userRes,
		Knives: knives,
	}

	s.renderTemplate(w, "user.html", payload)
}

func (s *Server) AdminIndex(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	knives, err := s.db.GetCollection(ctx)
	if err != nil {
		servererr(w, err, http.StatusInternalServerError)
		return
	}

	s.renderTemplate(w, "admin.html", struct {
		Knives []*db.KnifeType
	}{
		Knives: knives,
	})
}

func (s *Server) getUserForRequest(ctx context.Context, r *http.Request) (*db.User, error) {
	cookie, err := r.Cookie(AUTH_COOKIE)
	if err != nil {
		return nil, err
	}

	t, err := base64.RawStdEncoding.DecodeString(cookie.Value)
	if err != nil {
		return nil, err
	}

	auth, err := s.db.GetAuth(ctx, t)
	if err != nil {
		return nil, err
	}

	user, err := s.db.GetUserByID(ctx, auth.UserID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Server) KnifeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		servererr(w, err, http.StatusBadRequest)
		return
	}

	knife, err := s.db.GetKnife(ctx, id)
	if err != nil {
		servererr(w, err, http.StatusInternalServerError)
		return
	}

	payload := KnifePage{
		Knife:       *knife,
		RarityClass: className(knife.Rarity),
	}

	s.renderTemplate(w, "knife.html", payload)
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
		log.Printf("error parsing body %s", err)
		servererr(w, err, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	log.Printf("Operating on %+v", reqBody)

	var user *db.User
	if reqBody.TwitchID != "" {
		user, err = s.db.GetUserByTwitchID(ctx, reqBody.TwitchID)
	} else {
		user, err = s.db.GetUserByUsername(ctx, reqBody.Username)
	}
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Printf("pullHandler creating user")
			user, err = s.db.CreateUser(ctx, &db.User{
				Name:     reqBody.Username,
				TwitchID: reqBody.TwitchID,
			})
		}
		if err != nil {
			log.Printf("pullHandler getting user: %s", err)
			servererr(w, err, http.StatusBadRequest)
			return
		}
	}

	subscriber := reqBody.Subscriber == "1" || strings.ToLower(reqBody.Subscriber) == "true"
	verified := reqBody.Verified == "1" || strings.ToLower(reqBody.Verified) == "true"

	edition := 1
	edition, err = strconv.Atoi(reqBody.Edition)
	if err != nil {
		log.Printf("got edition %s and could not parse to number", reqBody.Edition)
		servererr(w, fmt.Errorf("unprocessable edition %s: %w", reqBody.Edition, err), http.StatusBadRequest)
		return
	}

	k, err := s.db.PullKnife(ctx, user.ID, reqBody.Knifename, subscriber, verified, edition)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Unable to find either this knife or this user: %s", err)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Internal Server Error: %s", err)
		return
	}

	err = json.NewEncoder(w).Encode(k)
	if err != nil {
		log.Printf("error serializing knife: %s", err)
	}
}

// Display the page that shows all the knives earnable
func (s *Server) CatalogHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	knives, err := s.db.GetCollection(ctx)
	if err != nil {
		servererr(w, err, http.StatusInternalServerError)
		return
	}
	var payload struct {
		Knives []struct {
			*db.KnifeType
			RarityClass string
		}
	}

	for _, k := range knives {
		payload.Knives = append(payload.Knives, struct {
			*db.KnifeType
			RarityClass string
		}{
			k,
			className(k.Rarity),
		})
	}

	s.renderTemplate(w, "catalog.html", payload)
}

// CatalogView renders a page showing a specific earnable knife
func (s *Server) CatalogView(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		servererr(w, err, http.StatusInternalServerError)
		return
	}

	knife, err := s.db.GetKnifeType(ctx, id)
	if err != nil {
		servererr(w, err, http.StatusInternalServerError)
		return
	}

	payload := struct {
		*db.KnifeType
		RarityClass string
	}{
		knife,
		className(knife.Rarity),
	}

	s.renderTemplate(w, "catalog-knife.html", payload)
}

// AdminKnifeList renders the admin page for viewing knives
func (s *Server) AdminKnifeList(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "NOT IMPLEMENTED", http.StatusInternalServerError)
}

// AdminKnife renders a page showing a current knife with a form to update
func (s *Server) AdminKnife(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		servererr(w, err, http.StatusInternalServerError)
		return
	}

	knife, err := s.db.GetKnifeType(ctx, id)
	if err != nil {
		servererr(w, err, http.StatusInternalServerError)
		return
	}

	payload := struct {
		*db.KnifeType
		RarityClass string
	}{
		knife,
		className(knife.Rarity),
	}

	s.renderTemplate(w, "admin-knife.html", payload)
}

// AdminUpdateKnife is the endpoint for modifying a knife
func (s *Server) AdminUpdateKnife(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse the multipart form
	err := r.ParseMultipartForm(32 << 20) // 32 MB maximum file size
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	author := r.FormValue("author")
	rarity := r.FormValue("rarity")
	basename := ""

	if r.Form.Has("image") {
		// Get the file from the request
		file, handler, err := r.FormFile("image")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Check the file type and size
		if !strings.HasPrefix(handler.Header.Get("Content-Type"), "image/") {
			http.Error(w, "File is not an image", http.StatusBadRequest)
			return
		}
		if handler.Size > 32<<20 {
			http.Error(w, "File is too large", http.StatusBadRequest)
			return
		}

		basename := path.Base(handler.Filename)
		if s.minioClient != nil {
			_, err = s.minioClient.PutObject(ctx, s.bucketName, path.Join("images", basename), file, handler.Size, minio.PutObjectOptions{
				ContentType: handler.Header.Get("Content-Type"),
			})
			if err != nil {
				http.Error(w, fmt.Sprintf("Error uploading image %s", err), http.StatusBadRequest)
				return
			}
		}
	}

	user, err := s.db.GetUserByUsername(ctx, author)
	if err != nil {
		http.Error(w, fmt.Sprintf("Unknown user %s", author), http.StatusBadRequest)
		return
	}

	// Save Knife finally to db
	k, err := s.db.UpdateKnifeType(ctx, &db.KnifeType{
		Name:      name,
		AuthorID:  user.ID,
		Rarity:    rarity,
		ImageName: basename,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to adminKnifePage for new knife
	http.Redirect(w, r, s.baseURL+"/admin/knife/"+strconv.Itoa(k.ID), http.StatusTemporaryRedirect)
}

func (s *Server) AdminDeleteKnife(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		servererr(w, err, http.StatusBadRequest)
		return
	}

	err = s.db.DeleteKnifeType(ctx, &db.KnifeType{
		ID: id,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("error deleting %d: %s", id, err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// AdminCreateKnife is the endpoint for creating a new new knife
func (s *Server) AdminCreateKnife(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse the multipart form
	err := r.ParseMultipartForm(32 << 20) // 32 MB maximum file size
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	author := r.FormValue("author")
	rarity := r.FormValue("rarity")

	// Get the file from the request
	file, handler, err := r.FormFile("image")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Check the file type and size
	if !strings.HasPrefix(handler.Header.Get("Content-Type"), "image/") {
		http.Error(w, "File is not an image", http.StatusBadRequest)
		return
	}
	if handler.Size > 32<<20 {
		http.Error(w, "File is too large", http.StatusBadRequest)
		return
	}

	user, err := s.db.GetUserByUsername(ctx, author)
	if err != nil {
		http.Error(w, fmt.Sprintf("Unknown user %s", author), http.StatusBadRequest)
		return
	}

	basename := path.Base(handler.Filename)
	if s.minioClient != nil {
		_, err = s.minioClient.PutObject(ctx, s.bucketName, path.Join("images", basename), file, handler.Size, minio.PutObjectOptions{
			ContentType: handler.Header.Get("Content-Type"),
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("Error uploading image %s", err), http.StatusBadRequest)
			return
		}
	}

	// Save Knife finally to db
	k, err := s.db.CreateKnifeType(ctx, &db.KnifeType{
		Name:      name,
		AuthorID:  user.ID,
		Rarity:    rarity,
		ImageName: basename,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to adminKnifePage for new knife
	http.Redirect(w, r, s.baseURL+"/admin/knife/"+strconv.Itoa(k.ID), http.StatusTemporaryRedirect)
}

func (s *Server) OnlyAdmin(inner http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Read the cookie for the user
		// Check if user is Admin
		// Then call inner otherwise 403
		user, err := s.getUserForRequest(r.Context(), r)
		if err != nil {
			servererr(w, err, http.StatusForbidden)
			return
		}

		if !user.Admin {
			servererr(w, fmt.Errorf("you, %s, are not authorized", user.Name), http.StatusForbidden)
			return
		}

		inner(w, r)
	}
}

func timeAgo(t time.Time) string {
	delta := time.Since(t)

	if delta < time.Minute {
		return "just now"
	}
	if delta < time.Hour {
		return fmt.Sprintf("%d minutes ago", int(delta.Minutes()))
	}
	if delta < 24*time.Hour {
		return fmt.Sprintf("%d hours ago", int(delta.Hours()))
	}
	return fmt.Sprintf("%d days ago", int(delta.Hours())/24)
}
