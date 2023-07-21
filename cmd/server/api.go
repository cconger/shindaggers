package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cconger/shindaggers/pkg/db"

	"github.com/gorilla/mux"
)

var errUnimplmeneted = fmt.Errorf("unimplemented")

type apierror struct {
	StatusCode   int    `json:"status"`
	ErrorMessage string `json:"message"`
	RequestID    string `json:"id"`
}

func serveAPIErr(w http.ResponseWriter, err error, statusCode int, userMessage string) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.WriteHeader(statusCode)

	if userMessage == "" {
		userMessage = "Unexpected Error"
	}

	// TODO(cconger): Proper telemetry here
	log.Printf("apierror (%d) %s: %s", statusCode, userMessage, err.Error())
	writeErr := json.NewEncoder(w).Encode(&apierror{
		StatusCode:   statusCode,
		ErrorMessage: userMessage,
		RequestID:    "", // TODO(cconger): extract created request id
	})

	if writeErr != nil {
		log.Printf("Unable to write apierror to responsewriter: %s", err)
	}
}

func serveAPIPayload(w http.ResponseWriter, payload interface{}) {
	w.Header().Add("Access-Control-Allow-Origin", "*")

	err := json.NewEncoder(w).Encode(payload)
	if err != nil {
		serveAPIErr(w, err, http.StatusInternalServerError, "Failed serializing response")
		return
	}
}

type IssuedCollectable struct {
	Collectable

	InstanceID string    `json:"instance_id"`
	Owner      User      `json:"owner"`
	Verified   bool      `json:"verified"`
	Subscriber bool      `json:"subscriber"`
	Edition    string    `json:"edition"`
	IssuedAt   time.Time `json:"issued_at"`
	Deleted    bool      `json:"deleted"`
}

func IssuedCollectableFromDBKnife(k *db.Knife) IssuedCollectable {
	res := IssuedCollectable{
		Collectable: CollectableFromDBKnife(k),
		InstanceID:  strconv.Itoa(k.InstanceID),
		Owner: User{
			ID:   strconv.Itoa(k.OwnerID),
			Name: k.Owner,
		},
		Verified:   k.Verified,
		Subscriber: k.Subscriber,
		Edition:    k.Edition,
		IssuedAt:   k.ObtainedAt,
		Deleted:    k.Deleted,
	}

	return res
}

type Collectable struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Author    User   `json:"author"`
	Rarity    string `json:"rarity"`
	ImagePath string `json:"image_path"`
}

func CollectableFromDBKnife(k *db.Knife) Collectable {
	return Collectable{
		ID:   strconv.Itoa(k.ID),
		Name: k.Name,
		Author: User{
			ID:   strconv.Itoa(k.AuthorID),
			Name: k.Author,
		},
		Rarity:    k.Rarity,
		ImagePath: k.ImageName,
	}
}

func CollectableFromDBKnifeType(k *db.KnifeType) Collectable {
	return Collectable{
		ID:   strconv.Itoa(k.ID),
		Name: k.Name,
		Author: User{
			ID:   strconv.Itoa(k.AuthorID),
			Name: k.Author,
		},
		Rarity:    k.Rarity,
		ImagePath: k.ImageName,
	}
}

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (s *Server) getIssuedCollectable(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		serveAPIErr(w, err, http.StatusBadRequest, "Could not parse id")
		return
	}

	knife, err := s.db.GetKnife(ctx, id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			serveAPIErr(w, err, http.StatusNotFound, "Unknown collectable")
			return
		}
		serveAPIErr(w, err, http.StatusInternalServerError, "")
		return
	}

	serveAPIPayload(
		w,
		&struct {
			Collectable IssuedCollectable
		}{
			Collectable: IssuedCollectableFromDBKnife(knife),
		},
	)
}

func (s *Server) getCollection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	dbknives, err := s.db.GetCollection(ctx, false)
	if err != nil {
		serveAPIErr(w, err, http.StatusInternalServerError, "")
		return
	}

	collectables := make([]Collectable, len(dbknives))

	for i, k := range dbknives {
		collectables[i] = CollectableFromDBKnifeType(k)
	}

	serveAPIPayload(
		w,
		&struct {
			Collectables []Collectable
		}{
			Collectables: collectables,
		},
	)
}

func (s *Server) getCollectable(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		serveAPIErr(w, err, http.StatusBadRequest, "Could not parse collectable id")
		return
	}

	dbknife, err := s.db.GetKnifeType(ctx, id, false)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			serveAPIErr(w, err, http.StatusNotFound, "Unknown collectable")
			return
		}
		serveAPIErr(w, err, http.StatusInternalServerError, "")
		return
	}

	serveAPIPayload(
		w,
		&struct {
			Collectable Collectable
		}{
			Collectable: CollectableFromDBKnifeType(dbknife),
		},
	)
}

func (s *Server) getLatest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// TODO: Query Param `since` for limiting time

	dbk, err := s.db.GetLatestPulls(ctx)
	if err != nil {
		serveAPIErr(w, err, http.StatusInternalServerError, "Internal Error")
		return
	}

	ic := make([]IssuedCollectable, len(dbk))

	for i, k := range dbk {
		ic[i] = IssuedCollectableFromDBKnife(k)
	}

	serveAPIPayload(
		w,
		&struct {
			Collectables []IssuedCollectable
		}{
			Collectables: ic,
		},
	)
}

func (s *Server) getLoggedInUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	rawToken := r.Header.Get("Authorization")
	t, err := base64.URLEncoding.DecodeString(rawToken)
	if err != nil {
		serveAPIErr(w, err, http.StatusForbidden, "Authorization token unreadable")
		return
	}

	auth, err := s.db.GetAuth(ctx, t)
	if err != nil {
		serveAPIErr(w, err, http.StatusForbidden, "Invalid Token")
		return
	}

	u, err := s.db.GetUserByID(ctx, auth.UserID)
	if err != nil {
		serveAPIErr(w, err, http.StatusForbidden, "Could not get user for token")
		return
	}

	serveAPIPayload(
		w,
		&struct {
			User User
		}{
			User: User{
				ID:   strconv.Itoa(u.ID),
				Name: u.Name,
			},
		},
	)
}

func (s *Server) getUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		serveAPIErr(w, err, http.StatusBadRequest, "Could not parse user id")
		return
	}

	u, err := s.db.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			serveAPIErr(w, err, http.StatusNotFound, "Unknown user")
			return
		}
		serveAPIErr(w, err, http.StatusInternalServerError, "")
		return
	}

	serveAPIPayload(
		w,
		&struct {
			User User
		}{
			User: User{
				ID:   strconv.Itoa(u.ID),
				Name: u.Name,
			},
		},
	)
}

func (s *Server) getUserByName(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)

	name := strings.ToLower(vars["name"])

	u, err := s.db.GetUserByUsername(ctx, name)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			serveAPIErr(w, err, http.StatusNotFound, "Unknown user")
			return
		}
		serveAPIErr(w, err, http.StatusInternalServerError, "")
		return
	}

	serveAPIPayload(
		w,
		&struct {
			User User
		}{
			User: User{
				ID:   strconv.Itoa(u.ID),
				Name: u.Name,
			},
		},
	)
}

func (s *Server) getUserCollectionByName(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)

	name, ok := vars["name"]

	if !ok {
		serveAPIErr(w, fmt.Errorf("id required"), http.StatusBadRequest, "User ID Required")
		return
	}

	// TODO: Check if userid is all numbers
	// If try to load it as an ID then twitchID
	// If not fallback to name

	userRes, err := s.db.GetUserByUsername(ctx, strings.ToLower(name))
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			serveAPIErr(w, err, http.StatusNotFound, "Unknown user")
			return
		}
		serveAPIErr(w, err, http.StatusInternalServerError, "")
		return
	}

	issuedRaw, err := s.db.GetKnivesForUsername(ctx, strings.ToLower(name))
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			serveAPIErr(w, err, http.StatusNotFound, "Unknown user")
			return
		}
		serveAPIErr(w, err, http.StatusInternalServerError, "")
		return
	}

	issuedCollectables := make([]IssuedCollectable, len(issuedRaw))
	for i, raw := range issuedRaw {
		issuedCollectables[i] = IssuedCollectableFromDBKnife(raw)
	}

	serveAPIPayload(
		w,
		&struct {
			User         User
			Collectables []IssuedCollectable
		}{
			User: User{
				ID:   strconv.Itoa(userRes.ID),
				Name: userRes.Name,
			},
			Collectables: issuedCollectables,
		},
	)
}

func (s *Server) getUserCollectionByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		serveAPIErr(w, fmt.Errorf("id required"), http.StatusBadRequest, "User ID Required")
		return
	}

	userRes, err := s.db.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			serveAPIErr(w, err, http.StatusNotFound, "Unknown user")
			return
		}
		serveAPIErr(w, err, http.StatusInternalServerError, "")
		return
	}

	issuedRaw, err := s.db.GetKnivesForUsername(ctx, userRes.LookupName)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			serveAPIErr(w, err, http.StatusNotFound, "Unknown user")
			return
		}
		serveAPIErr(w, err, http.StatusInternalServerError, "")
		return
	}

	issuedCollectables := make([]IssuedCollectable, len(issuedRaw))
	for i, raw := range issuedRaw {
		issuedCollectables[i] = IssuedCollectableFromDBKnife(raw)
	}

	serveAPIPayload(
		w,
		&struct {
			User         User
			Collectables []IssuedCollectable
		}{
			User: User{
				ID:   strconv.Itoa(userRes.ID),
				Name: userRes.Name,
			},
			Collectables: issuedCollectables,
		},
	)
}
