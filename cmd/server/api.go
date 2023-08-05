package main

import (
	"context"
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

func (s *Server) getAuth(ctx context.Context, r *http.Request) (*db.UserAuth, error) {
	rawToken := r.Header.Get("Authorization")
	t, err := base64.URLEncoding.DecodeString(rawToken)
	if err != nil {
		return nil, fmt.Errorf("authorization token unreadable")
	}

	auth, err := s.db.GetAuth(ctx, t)
	if err != nil {
		return nil, fmt.Errorf("invalid token")
	}

	return auth, nil
}

func (s *Server) getLoggedInUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	auth, err := s.getAuth(ctx, r)
	if err != nil {
		serveAPIErr(w, err, http.StatusForbidden, "Unable to identify authenticated user")
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

	useridstr, ok := vars["userid"]
	if !ok {
		serveAPIErr(w, fmt.Errorf("id required"), http.StatusBadRequest, "User ID Required")
		return
	}

	user, err := s.getUserByUserID(ctx, ParseUserID(useridstr))
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
				ID:   strconv.Itoa(user.ID),
				Name: user.Name,
			},
		},
	)
}

func (s *Server) getEquippedForUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)

	useridstr, ok := vars["userid"]
	if !ok {
		serveAPIErr(w, fmt.Errorf("id required"), http.StatusBadRequest, "User ID Required")
		return
	}

	user, err := s.getUserByUserID(ctx, ParseUserID(useridstr))
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			serveAPIErr(w, err, http.StatusNotFound, "Unknown user")
			return
		}
		serveAPIErr(w, err, http.StatusInternalServerError, "")
		return
	}

	eqRaw, err := s.db.GetEquippedKnifeForUser(ctx, user.ID)
	if err != nil {
		serveAPIErr(w, err, http.StatusInternalServerError, "Unable to fetch equipped knife")
		return
	}
	var equipped *IssuedCollectable
	if eqRaw != nil {
		eq := IssuedCollectableFromDBKnife(eqRaw)
		equipped = &eq
	}

	serveAPIPayload(
		w,
		&struct {
			User     User
			Equipped *IssuedCollectable
		}{
			User: User{
				ID:   strconv.Itoa(user.ID),
				Name: user.Name,
			},
			Equipped: equipped,
		},
	)
}

func (s *Server) getUserByUserID(ctx context.Context, userID UserID) (*db.User, error) {
	if userID.IsTwitch() {
		return s.db.GetUserByTwitchID(ctx, userID.TwitchID)
	} else if userID.IsInternal() {
		return s.db.GetUserByID(ctx, userID.InternalID)
	}
	return s.db.GetUserByUsername(ctx, userID.Name)
}

func (s *Server) getUserCollection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)

	useridstr, ok := vars["userid"]
	if !ok {
		serveAPIErr(w, fmt.Errorf("id required"), http.StatusBadRequest, "User ID Required")
		return
	}

	user, err := s.getUserByUserID(ctx, ParseUserID(useridstr))
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			serveAPIErr(w, err, http.StatusNotFound, "Unknown user")
			return
		}
		serveAPIErr(w, err, http.StatusInternalServerError, "")
		return
	}

	issuedRaw, err := s.db.GetKnivesForUsername(ctx, strings.ToLower(user.LookupName))
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

	eqRaw, err := s.db.GetEquippedKnifeForUser(ctx, user.ID)
	if err != nil {
		serveAPIErr(w, err, http.StatusInternalServerError, "Unable to fetch equipped knife")
		return
	}
	var equipped *IssuedCollectable
	if eqRaw != nil {
		eq := IssuedCollectableFromDBKnife(eqRaw)
		equipped = &eq
	}

	serveAPIPayload(
		w,
		&struct {
			User         User
			Collectables []IssuedCollectable
			Equipped     *IssuedCollectable
		}{
			User: User{
				ID:   strconv.Itoa(user.ID),
				Name: user.Name,
			},
			Collectables: issuedCollectables,
			Equipped:     equipped,
		},
	)
}

type EquipPayload struct {
	UserID   string
	IssuedID string
}

func (s *Server) EquipHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	a, err := s.getAuth(ctx, r)
	if err != nil {
		serveAPIErr(w, err, http.StatusForbidden, "Unable to identify authenticated user")
		return
	}

	var payload EquipPayload
	err = json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		serveAPIErr(w, err, http.StatusBadRequest, "Could not parse body")
		return
	}

	if payload.UserID == "" {
		serveAPIErr(w, fmt.Errorf("payload has zero value for UserID"), http.StatusBadRequest, "UserID must be specified")
		return
	}

	parseduid, err := strconv.Atoi(payload.UserID)
	if err != nil {
		serveAPIErr(w, err, http.StatusBadRequest, "UserID not numeric ")
		return
	}

	if payload.IssuedID == "" {
		serveAPIErr(w, fmt.Errorf("payload has zero value for IssuedID"), http.StatusBadRequest, "InstanceID must be specified")
		return
	}

	issuedID, err := strconv.Atoi(payload.IssuedID)
	if err != nil {
		serveAPIErr(w, err, http.StatusBadRequest, "IssuedID not numeric ")
		return
	}

	// Lookup user
	user, err := s.db.GetUserByID(ctx, a.UserID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			serveAPIErr(w, err, http.StatusNotFound, "Unknown user")
			return
		}
		serveAPIErr(w, err, http.StatusInternalServerError, "")
		return
	}

	if !user.Admin && (user.ID != parseduid) {
		serveAPIErr(
			w,
			fmt.Errorf("non admin user (%d) tried to equip knife for someone else", user.ID),
			http.StatusForbidden,
			"You cannot equip knives for other users",
		)
		return
	}

	// Lookup if knife is owned by user
	issuedRaw, err := s.db.GetKnife(ctx, issuedID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			serveAPIErr(w, err, http.StatusNotFound, "Unknown issued collectable")
			return
		}
		serveAPIErr(w, err, http.StatusInternalServerError, "")
		return
	}

	if issuedRaw.OwnerID != parseduid {
		serveAPIErr(
			w,
			fmt.Errorf("user doesn't own collectable requested to equip"),
			http.StatusBadRequest,
			"Specified user does not own the collectable specified",
		)
		return
	}

	err = s.db.EquipKnifeForUser(ctx, parseduid, issuedID)
	if err != nil {
		serveAPIErr(w, err, http.StatusInternalServerError, "")
		return
	}
}
