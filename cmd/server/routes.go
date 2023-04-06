package main

import (
	"crypto/rand"
	"crypto/sha256"
	"embed"
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
)

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
}

func (s *Server) getTemplate(templateName string) (*template.Template, error) {
	if s.devMode {
		return template.ParseFiles(path.Join("cmd", "server", "templates", templateName))
	}
	return template.ParseFS(templates, path.Join("templates", templateName))
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

	t, err := s.getTemplate("index.html")
	if err != nil {
		servererr(w, err, http.StatusInternalServerError)
		return
	}

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

	w.WriteHeader(http.StatusOK)
	t.Execute(w, pl)
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
	user, err := s.db.GetUser(ctx, twitchUser.Login)
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

	userRes, err := s.db.GetUser(ctx, username)
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

	t, err := s.getTemplate("user.html")
	if err != nil {
		servererr(w, err, http.StatusInternalServerError)
		return
	}

	payload := UserPagePayload{
		User:   userRes,
		Knives: knives,
	}

	err = t.Execute(w, payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("error executing template: %s", err)
	}
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

	t, err := s.getTemplate("knife.html")
	if err != nil {
		servererr(w, err, http.StatusInternalServerError)
		return
	}

	payload := KnifePage{
		Knife:       *knife,
		RarityClass: className(knife.Rarity),
	}

	err = t.Execute(w, payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("error executing template: %s", err)
	}
}

type PullRequest struct {
	TwitchID  string `json:"twitch_id"`
	Username  string `json:"username"`
	Knifename string `json:"knifename"`
}

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

	// TODO: lookup user by id or username here instead of in db
	k, err := s.db.PullKnife(ctx, reqBody.Username, reqBody.Knifename)
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

	t, err := s.getTemplate("catalog.html")
	if err != nil {
		servererr(w, err, http.StatusInternalServerError)
		return
	}

	t.Execute(w, payload)
}

// Display the page that shows all the knives earnable
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

	t, err := s.getTemplate("catalog-knife.html")
	if err != nil {
		servererr(w, err, http.StatusInternalServerError)
		return
	}

	t.Execute(w, payload)
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
