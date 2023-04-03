package main

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/cconger/shindaggers/pkg/db"

	"github.com/gorilla/mux"
)

type FrontPage struct{}

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

//go:embed templates/*
var templates embed.FS

func servererr(w http.ResponseWriter, err error, errorCode int) {
	w.WriteHeader(errorCode)
	fmt.Fprintf(w, "Error: %s", err)
}

type Server struct {
	devMode bool
	db      db.KnifeDB
}

func (s *Server) getTemplate(templateName string) (*template.Template, error) {
	if s.devMode {
		return template.ParseFiles(path.Join("cmd", "server", "templates", templateName))
	}
	return template.ParseFS(templates, path.Join("templates", templateName))
}

func (s *Server) IndexHandler(w http.ResponseWriter, r *http.Request) {
	t, err := s.getTemplate("index.html")
	if err != nil {
		servererr(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	t.Execute(w, nil)
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
		log.Println("error executing template: %s", err)
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
		log.Println("error executing template: %s", err)
	}
}
