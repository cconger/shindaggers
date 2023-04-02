package main

import (
	"embed"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/cconger/shindaggers/pkg/db"

	"github.com/gorilla/mux"
)

type FrontPage struct{}

//go:embed templates/*
var templates embed.FS

func servererr(w http.ResponseWriter, err error, errorCode int) {
	w.WriteHeader(errorCode)
	fmt.Fprintf(w, "Error: %s", err)
}

func index(w http.ResponseWriter, r *http.Request) {
	f, err := templates.Open("templates/index.html")
	if err != nil {
		servererr(w, err, http.StatusInternalServerError)
		return
	}
	defer f.Close()

	w.WriteHeader(http.StatusOK)
	io.Copy(w, f)
}

type Server struct {
	db db.KnifeDB
}

func (s *Server) UserHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "ENDPOINT NOT IMPLEMENTED")
}

func (s *Server) KnifeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ctx := r.Context()

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		servererr(w, err, http.StatusBadRequest)
	}

	knife, err := s.db.GetKnife(ctx, id)
	if err != nil {
		servererr(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Knife: %s", knife.Name)
}
