package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/cconger/shindaggers/pkg/db"
	model "github.com/cconger/shindaggers/pkg/db/.gen/postgres/public/model"

	"github.com/gorilla/mux"
)

var (
	errMissingField  = fmt.Errorf("missing required field")
	errUnimplmeneted = fmt.Errorf("unimplemented")
	errAdminOnly     = fmt.Errorf("admin only")
)

const (
	RarityCommon    = "Common"
	RarityUncommon  = "Uncommon"
	RarityRare      = "Rare"
	RaritySuperRare = "Super Rare"
	RarityUltraRare = "Ultra Rare"
)

var rarities = []string{
	RarityCommon,
	RarityUncommon,
	RarityRare,
	RaritySuperRare,
	RarityUltraRare,
}

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

	slog.Error("apierror", "statuscode", statusCode, "userMessage", userMessage, "err", err)

	writeErr := json.NewEncoder(w).Encode(&apierror{
		StatusCode:   statusCode,
		ErrorMessage: userMessage,
		RequestID:    "", // TODO(cconger): extract created request id
	})

	if writeErr != nil {
		slog.Error("writing apierror", "err", err)
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
		InstanceID:  strconv.FormatInt(k.InstanceID, 10),
		Owner: User{
			ID:   strconv.FormatInt(k.OwnerID, 10),
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

type Tags struct {
	Verified   bool `json:"verified"`
	Subscriber bool `json:"subscriber"`
}

func IssuedCollectableFromCollectableInstance(i *db.CollectableInstance) IssuedCollectable {
	t := Tags{}
	if i.Tags != nil {
		err := json.Unmarshal([]byte(*i.Tags), &t)
		if err != nil {
			slog.Error("unable to unmarshal tags", "err", err, "tags", i.Tags)
		}
	}

	owner := User{
		ID: strconv.FormatInt(i.OwnerID, 10),
	}
	if i.Owner != nil {
		owner.ID = strconv.FormatInt(i.Owner.ID, 10)
		owner.Name = i.Owner.Name
	}
	edition := ""
	if i.Edition != nil {
		edition = i.Edition.Name
	}

	res := IssuedCollectable{
		Collectable: CollectableFromDBCollectable(i.Collectable),
		InstanceID:  strconv.FormatInt(i.ID, 10),
		Owner:       owner,
		Verified:    t.Verified,
		Subscriber:  t.Subscriber,
		Edition:     edition,
		IssuedAt:    i.CreatedAt,
		Deleted:     i.DeletedAt != nil,
	}
	return res
}

func CollectableFromDBCollectable(c *db.Collectable) Collectable {
	res := Collectable{
		ID:   strconv.FormatInt(c.ID, 10),
		Name: c.Name,
		Author: User{
			ID:   strconv.FormatInt(c.Creator.ID, 10),
			Name: c.Creator.Name,
		},
		Rarity:    c.Rarity,
		ImagePath: c.Imagepath,
		ImageURL:  "https://images.shindaggers.io/images/" + c.Imagepath,
	}
	return res
}

func IssuedCollectableFromDBCollectable(c *db.Collectable) IssuedCollectable {
	res := IssuedCollectable{
		Collectable: CollectableFromDBCollectable(c),
	}

	return res
}

type Collectable struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Author    User   `json:"author"`
	Rarity    string `json:"rarity"`
	ImagePath string `json:"image_path"`
	ImageURL  string `json:"image_url"`
}

func CollectableFromDBKnife(k *db.Knife) Collectable {
	return Collectable{
		ID:   strconv.FormatInt(k.ID, 10),
		Name: k.Name,
		Author: User{
			ID:   strconv.FormatInt(k.AuthorID, 10),
			Name: k.Author,
		},
		Rarity:    k.Rarity,
		ImagePath: k.ImageName,
		ImageURL:  "https://images.shindaggers.io/images/" + k.ImageName,
	}
}

func CollectableFromDBKnifeType(k *db.KnifeType) Collectable {
	return Collectable{
		ID:   strconv.FormatInt(k.ID, 10),
		Name: k.Name,
		Author: User{
			ID:   strconv.FormatInt(k.AuthorID, 10),
			Name: k.Author,
		},
		Rarity:    k.Rarity,
		ImagePath: k.ImageName,
		ImageURL:  "https://images.shindaggers.io/images/" + k.ImageName,
	}
}

type AdminCollectable struct {
	Collectable

	Deleted  bool `json:"deleted"`
	Approved bool `json:"approved"`
}

func AdminCollectableFromDBCollectable(k *db.Collectable) AdminCollectable {
	return AdminCollectable{
		Collectable: CollectableFromDBCollectable(k),
		Deleted:     k.DeletedAt != nil,
		Approved:    k.ApprovedAt != nil,
	}
}

func UserFromDBUser(u *db.User) User {
	return User{
		ID:   strconv.FormatInt(u.ID, 10),
		Name: u.Name,
	}
}

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type UserStats struct {
	Wins   int `json:"wins"`
	Losses int `json:"losses"`
	Ties   int `json:"ties"`
}

func (s *Server) getIssuedCollectable(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)

	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		serveAPIErr(w, err, http.StatusBadRequest, "Could not parse id")
		return
	}

	c, err := s.db.GetCollectableInstances(ctx, db.GetCollectableInstancesOptions{
		ByID: id,
	})
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			serveAPIErr(w, err, http.StatusNotFound, "Unknown collectable")
			return
		}
		serveAPIErr(w, err, http.StatusInternalServerError, "")
		return
	}

	if len(c) == 0 {
		serveAPIErr(w, err, http.StatusNotFound, "Unknown collectable")
		return
	}

	res := IssuedCollectableFromCollectableInstance(&c[0])

	serveAPIPayload(
		w,
		&res,
	)
}

func (s *Server) getCollection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	c, err := s.db.GetCollectables(ctx, db.GetCollectablesOptions{
		Collection: 1,
	})
	if err != nil {
		serveAPIErr(w, err, http.StatusInternalServerError, "")
		return
	}

	res := make([]Collectable, len(c))
	for i, c := range c {
		res[i] = CollectableFromDBCollectable(c)
	}

	serveAPIPayload(
		w,
		&res,
	)
}

func (s *Server) getCollectable(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)

	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		serveAPIErr(w, err, http.StatusBadRequest, "Could not parse collectable id")
		return
	}

	c, err := s.db.GetCollectable(ctx, id, db.GetCollectableOptions{})
	if err != nil {
		serveAPIErr(w, err, http.StatusInternalServerError, "")
		return
	}

	res := CollectableFromDBCollectable(c)

	serveAPIPayload(
		w,
		&res,
	)
}

func (s *Server) getLatest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	options := db.GetLatestIssuesOptions{
		ByCollection: 1,
	}

	since := r.URL.Query().Get("since")
	if since != "" {
		ms, err := strconv.ParseInt(since, 10, 64)
		if err != nil {
			serveAPIErr(w, err, http.StatusBadRequest, "since not encoded properly")
			return
		}
		options.After = time.UnixMilli(ms)
	}

	lp, err := s.db.GetLatestIssues(ctx, options)
	if err != nil {
		serveAPIErr(w, err, http.StatusInternalServerError, "Internal Error")
		return
	}

	res := make([]IssuedCollectable, len(lp))
	for i, c := range lp {
		res[i] = IssuedCollectableFromCollectableInstance(&c)
	}

	serveAPIPayload(
		w,
		&res,
	)
}

func (s *Server) getAuthUser(ctx context.Context, r *http.Request) (*db.User, error) {
	rawToken := r.Header.Get("Authorization")
	t, err := base64.URLEncoding.DecodeString(rawToken)
	if err != nil {
		return nil, fmt.Errorf("authorization token unreadable")
	}

	u, err := s.db.GetUser(ctx, db.GetUserOptions{
		AuthToken: t,
	})
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (s *Server) getLoggedInUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	u, err := s.getAuthUser(ctx, r)
	if err != nil {
		serveAPIErr(w, err, http.StatusForbidden, "Could not get user for token")
		return
	}

	serveAPIPayload(
		w,
		&struct {
			User User
		}{
			User: UserFromDBUser(u),
		},
	)
}

func (s *Server) getUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	search := r.URL.Query().Get("search")
	if search == "" {
		serveAPIErr(
			w,
			fmt.Errorf("search query for getUsers required"),
			http.StatusBadRequest,
			"Search param required",
		)
		return
	}

	udbs, err := s.db.SearchUsers(ctx, search)
	if err != nil {
		serveAPIErr(
			w,
			err,
			http.StatusInternalServerError,
			"error loading users",
		)
		return
	}

	users := make([]User, len(udbs))
	for i, u := range udbs {
		users[i] = UserFromDBUser(&u)
	}

	serveAPIPayload(
		w,
		&struct {
			Users []User
		}{
			Users: users,
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
				ID:   strconv.FormatInt(user.ID, 10),
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

			c, err := s.getRandomCollectable(ctx)
			if err != nil {
				serveAPIErr(w, err, http.StatusInternalServerError, "Could not get random knife for user")
				return
			}
			slog.Warn("Creating a fake user response for a name that shindigs prolly made up")

			serveAPIPayload(
				w,
				&struct {
					User           User
					Equipped       *IssuedCollectable
					RandomlyPicked bool
					LoanerKnife    bool
					FakeUser       bool
				}{
					User: User{
						ID:   useridstr,
						Name: useridstr,
					},
					Equipped: &IssuedCollectable{
						Collectable: CollectableFromDBCollectable(c),
					},
					RandomlyPicked: true,
					LoanerKnife:    true,
					FakeUser:       true,
				},
			)
			return
		}
		serveAPIErr(w, err, http.StatusInternalServerError, "")
		return
	}

	eqRaw, err := s.db.GetEquippedForUser(ctx, user.ID)
	if err != nil {
		serveAPIErr(w, err, http.StatusInternalServerError, "Unable to fetch equipped knife")
		return
	}

	var equipped *IssuedCollectable
	var randomlypicked bool
	var loanerKnife bool
	if eqRaw == nil {
		raw, err := s.db.GetCollectableInstances(ctx, db.GetCollectableInstancesOptions{
			ByOwner: user.ID,
		})
		if err != nil {
			slog.Error("unable to get knives for user", "err", err, "user.id", user.ID)
		} else {
			if len(raw) > 0 {
				eq := IssuedCollectableFromCollectableInstance(&raw[rand.Intn(len(raw))])
				equipped = &eq
				randomlypicked = true
			} else {
				c, err := s.getRandomCollectable(ctx)
				if err != nil {
					serveAPIErr(w, err, http.StatusInternalServerError, "Could not fetch collection for loaner")
					return
				}

				equipped = &IssuedCollectable{
					Collectable: CollectableFromDBCollectable(c),
				}
				randomlypicked = true
				loanerKnife = true
			}
		}
	} else {
		eq := IssuedCollectableFromCollectableInstance(eqRaw)
		equipped = &eq
	}

	serveAPIPayload(
		w,
		&struct {
			User           User
			Equipped       *IssuedCollectable
			RandomlyPicked bool
			LoanerKnife    bool
		}{
			User: User{
				ID:   strconv.FormatInt(user.ID, 10),
				Name: user.Name,
			},
			Equipped:       equipped,
			RandomlyPicked: randomlypicked,
			LoanerKnife:    loanerKnife,
		},
	)
}

func (s *Server) getUserByUserID(ctx context.Context, userID UserID) (*db.User, error) {
	options := db.GetUserOptions{}
	if userID.IsTwitch() {
		options.TwitchID = userID.TwitchID
	} else if userID.IsInternal() {
		options.ID = userID.InternalID
	} else {
		options.Username = userID.Name
	}
	return s.db.GetUser(ctx, options)
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

	issuedRaw, err := s.db.GetCollectableInstances(ctx, db.GetCollectableInstancesOptions{
		ByOwner: user.ID,
	})
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
		issuedCollectables[i] = IssuedCollectableFromCollectableInstance(&raw)
	}

	eqRaw, err := s.db.GetEquippedForUser(ctx, user.ID)
	if err != nil {
		serveAPIErr(w, err, http.StatusInternalServerError, "Unable to fetch equipped knife")
		return
	}
	var equipped *IssuedCollectable
	if eqRaw != nil {
		eq := IssuedCollectableFromCollectableInstance(eqRaw)
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
				ID:   strconv.FormatInt(user.ID, 10),
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

	user, err := s.getAuthUser(ctx, r)
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

	parseduid, err := strconv.ParseInt(payload.UserID, 10, 64)
	if err != nil {
		serveAPIErr(w, err, http.StatusBadRequest, "UserID not numeric ")
		return
	}

	if payload.IssuedID == "" {
		serveAPIErr(w, fmt.Errorf("payload has zero value for IssuedID"), http.StatusBadRequest, "InstanceID must be specified")
		return
	}

	issuedID, err := strconv.ParseInt(payload.IssuedID, 10, 64)
	if err != nil {
		serveAPIErr(w, err, http.StatusBadRequest, "IssuedID not numeric ")
		return
	}

	if (user.Admin == nil || !*user.Admin) && (user.ID != parseduid) {
		serveAPIErr(
			w,
			fmt.Errorf("non admin user (%d) tried to equip knife for someone else", user.ID),
			http.StatusForbidden,
			"You cannot equip knives for other users",
		)
		return
	}

	// Lookup if knife is owned by user
	issuedRaw, err := s.db.GetCollectableInstances(ctx, db.GetCollectableInstancesOptions{
		ByID: issuedID,
	})
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			serveAPIErr(w, err, http.StatusNotFound, "Unknown issued collectable")
			return
		}
		serveAPIErr(w, err, http.StatusInternalServerError, "")
		return
	}
	if len(issuedRaw) == 0 {
		serveAPIErr(w, err, http.StatusNotFound, "Unknown issued collectable")
		return
	}

	if issuedRaw[0].OwnerID != parseduid {
		serveAPIErr(
			w,
			fmt.Errorf("user doesn't own collectable requested to equip"),
			http.StatusBadRequest,
			"Specified user does not own the collectable specified",
		)
		return
	}

	err = s.db.SetEquipped(ctx, issuedID, parseduid)
	if err != nil {
		serveAPIErr(w, err, http.StatusInternalServerError, "")
		return
	}
}

func (s *Server) adminListCollectables(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	u, err := s.getAuthUser(ctx, r)
	if err != nil {
		serveAPIErr(w, err, http.StatusForbidden, "could not identify user")
		return
	}

	if u.Admin == nil || !*u.Admin {
		serveAPIErr(w, errAdminOnly, http.StatusForbidden, "")
		return
	}

	dbknives, err := s.db.GetCollectables(ctx, db.GetCollectablesOptions{
		GetDeleted: true,
	})
	if err != nil {
		serveAPIErr(w, err, http.StatusInternalServerError, "")
		return
	}

	collectables := make([]AdminCollectable, len(dbknives))

	for i, k := range dbknives {
		collectables[i] = AdminCollectableFromDBCollectable(k)
	}

	pendingknives, err := s.db.GetCollectables(ctx, db.GetCollectablesOptions{
		GetUnapproved:  true,
		OnlyUnapproved: true,
	})
	if err != nil {
		serveAPIErr(w, err, http.StatusInternalServerError, "")
		return
	}

	pendingApproval := make([]AdminCollectable, len(pendingknives))
	for i, k := range pendingknives {
		pendingApproval[i] = AdminCollectableFromDBCollectable(k)
	}

	serveAPIPayload(
		w,
		&struct {
			ApprovalQueue []AdminCollectable
			Collectables  []AdminCollectable
		}{
			ApprovalQueue: pendingApproval,
			Collectables:  collectables,
		},
	)
}

type CollectablePayload struct {
	Collectable Collectable
}

func (s Server) createCollectable(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	u, err := s.getAuthUser(ctx, r)
	if err != nil {
		serveAPIErr(w, err, http.StatusForbidden, "could not identify user")
		return
	}

	var payload CollectablePayload
	err = json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		serveAPIErr(w, err, http.StatusBadRequest, "could not parse body")
		return
	}
	r.Body.Close()

	if payload.Collectable.Name == "" {
		serveAPIErr(w, errMissingField, http.StatusBadRequest, "Name cannot be empty")
		return
	}

	if !slices.Contains(rarities, payload.Collectable.Rarity) {
		serveAPIErr(w, errMissingField, http.StatusBadRequest, "Rarity is unknown")
		return
	}

	if payload.Collectable.ImagePath == "" {
		serveAPIErr(w, errMissingField, http.StatusBadRequest, "ImagePath cannot be empty")
		return
	}

	collectionID := int64(1)
	created, err := s.db.CreateCollectable(ctx, model.Collectables{
		ID:           s.idGenerator.Generate().Int64(),
		CollectionID: &collectionID,
		Name:         payload.Collectable.Name,
		CreatorID:    u.ID,
		Rarity:       payload.Collectable.Rarity,
		Imagepath:    payload.Collectable.ImagePath,
	})
	if err != nil {
		serveAPIErr(w, err, http.StatusInternalServerError, "could not create collectable")
		return
	}

	serveAPIPayload(w, struct {
		Collectable AdminCollectable
	}{
		Collectable: AdminCollectableFromDBCollectable(created),
	})
}

func (s *Server) adminCreateCollectable(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	u, err := s.getAuthUser(ctx, r)
	if err != nil {
		serveAPIErr(w, err, http.StatusForbidden, "could not identify user")
		return
	}

	if u.Admin == nil || !*u.Admin {
		serveAPIErr(w, errAdminOnly, http.StatusForbidden, "")
		return
	}

	var payload CollectablePayload
	err = json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		serveAPIErr(w, err, http.StatusBadRequest, "could not parse body")
		return
	}
	r.Body.Close()

	if payload.Collectable.Name == "" {
		serveAPIErr(w, errMissingField, http.StatusBadRequest, "Name cannot be empty")
		return
	}

	if payload.Collectable.Author.ID == "" {
		serveAPIErr(w, errMissingField, http.StatusBadRequest, "Author.ID cannot be empty")
		return
	}

	authorID, err := strconv.ParseInt(payload.Collectable.Author.ID, 10, 64)
	if err != nil {
		serveAPIErr(w, err, http.StatusBadRequest, "Author.ID is not parseable")
		return
	}

	if !slices.Contains(rarities, payload.Collectable.Rarity) {
		serveAPIErr(w, errMissingField, http.StatusBadRequest, "Rarity is unknown")
		return
	}

	if payload.Collectable.ImagePath == "" {
		serveAPIErr(w, errMissingField, http.StatusBadRequest, "ImagePath cannot be empty")
		return
	}

	collectionID := int64(1)
	now := time.Now()
	created, err := s.db.CreateCollectable(ctx, model.Collectables{
		ID:           s.idGenerator.Generate().Int64(),
		CollectionID: &collectionID,
		Name:         payload.Collectable.Name,
		CreatorID:    authorID,
		Rarity:       payload.Collectable.Rarity,
		Imagepath:    payload.Collectable.ImagePath,
		ApprovedAt:   &now,
		ApprovedBy:   &u.ID,
	})
	if err != nil {
		serveAPIErr(w, err, http.StatusInternalServerError, "could not create collectable")
		return
	}

	serveAPIPayload(w, struct {
		Collectable AdminCollectable
	}{
		Collectable: AdminCollectableFromDBCollectable(created),
	})
}

func (s *Server) adminDeleteCollectable(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	u, err := s.getAuthUser(ctx, r)
	if err != nil {
		serveAPIErr(w, err, http.StatusForbidden, "could not identify user")
		return
	}

	if u.Admin == nil || !*u.Admin {
		serveAPIErr(w, errAdminOnly, http.StatusForbidden, "")
		return
	}

	vars := mux.Vars(r)

	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		serveAPIErr(w, err, http.StatusBadRequest, "id is not numeric")
		return
	}

	err = s.db.DeleteCollectable(ctx, id)
	if err != nil {
		serveAPIErr(w, err, http.StatusInternalServerError, "could not delete knife")
		return
	}

	serveAPIPayload(w, true)
}

func (s *Server) adminUpdateCollectable(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	u, err := s.getAuthUser(ctx, r)
	if err != nil {
		serveAPIErr(w, err, http.StatusForbidden, "could not identify user")
		return
	}

	if u.Admin == nil || !*u.Admin {
		serveAPIErr(w, errAdminOnly, http.StatusForbidden, "")
		return
	}

	vars := mux.Vars(r)

	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		serveAPIErr(w, err, http.StatusBadRequest, "id is not numeric")
		return
	}

	var payload CollectablePayload
	err = json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		serveAPIErr(w, err, http.StatusBadRequest, "could not parse body")
		return
	}
	r.Body.Close()

	if payload.Collectable.Name == "" {
		serveAPIErr(w, errMissingField, http.StatusBadRequest, "Name cannot be empty")
		return
	}

	if payload.Collectable.Author.ID == "" {
		serveAPIErr(w, errMissingField, http.StatusBadRequest, "Author.ID cannot be empty")
		return
	}

	authorID, err := strconv.ParseInt(payload.Collectable.Author.ID, 10, 64)
	if err != nil {
		serveAPIErr(w, err, http.StatusBadRequest, "Author.ID is not parseable")
		return
	}

	if !slices.Contains(rarities, payload.Collectable.Rarity) {
		serveAPIErr(w, errMissingField, http.StatusBadRequest, "Rarity is unknown")
		return
	}

	if payload.Collectable.ImagePath == "" {
		serveAPIErr(w, errMissingField, http.StatusBadRequest, "ImagePath cannot be empty")
		return
	}

	created, err := s.db.UpdateCollectable(ctx, model.Collectables{
		ID:        id,
		Name:      payload.Collectable.Name,
		CreatorID: authorID,
		Rarity:    payload.Collectable.Rarity,
		Imagepath: payload.Collectable.ImagePath,
	})
	if err != nil {
		serveAPIErr(w, err, http.StatusInternalServerError, "could not create collectable")
		return
	}

	serveAPIPayload(w, struct {
		Collectable AdminCollectable
	}{
		Collectable: AdminCollectableFromDBCollectable(created),
	})
}

func (s *Server) adminIssueCollectable(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	u, err := s.getAuthUser(ctx, r)
	if err != nil {
		serveAPIErr(w, err, http.StatusForbidden, "could not identify user")
		return
	}

	if u.Admin == nil || !*u.Admin {
		serveAPIErr(w, errAdminOnly, http.StatusForbidden, "")
		return
	}

	serveAPIErr(w, fmt.Errorf("not implemented"), http.StatusNotImplemented, "Not Implemented")
}

func (s *Server) adminRevokeIssuedCollectable(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	u, err := s.getAuthUser(ctx, r)
	if err != nil {
		serveAPIErr(w, err, http.StatusForbidden, "could not identify user")
		return
	}

	if u.Admin == nil || !*u.Admin {
		serveAPIErr(w, errAdminOnly, http.StatusForbidden, "")
		return
	}

	serveAPIErr(w, fmt.Errorf("not implemented"), http.StatusNotImplemented, "Not Implemented")
}

type IssuedConfig struct {
	Weights map[string]int
}

func (s *Server) adminGetIssueConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	u, err := s.getAuthUser(ctx, r)
	if err != nil {
		serveAPIErr(w, err, http.StatusForbidden, "could not identify user")
		return
	}

	if u.Admin == nil || !*u.Admin {
		serveAPIErr(w, errAdminOnly, http.StatusForbidden, "")
		return
	}

	pullWeight, err := s.db.GetWeights(ctx, 1)
	if err != nil {
		serveAPIErr(w, err, http.StatusInternalServerError, "")
	}

	res := make(map[string]int)
	for _, w := range pullWeight {
		res[w.Rarity] = w.Weight
	}

	serveAPIPayload(w, IssuedConfig{Weights: res})
}

func (s *Server) adminUpdateIssueConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	u, err := s.getAuthUser(ctx, r)
	if err != nil {
		serveAPIErr(w, err, http.StatusForbidden, "could not identify user")
		return
	}

	if u.Admin == nil || !*u.Admin {
		serveAPIErr(w, errAdminOnly, http.StatusForbidden, "")
		return
	}

	serveAPIErr(w, fmt.Errorf("not implemented"), http.StatusNotImplemented, "Not Implemented")
}

type RandomPullRequest struct {
	TwitchID   string `json:"twitch_id"`
	Subscriber bool   `json:"subscriber"`
	DryRun     bool   `json:"dry_run"`
}

func (s *Server) RandomPullHandler(w http.ResponseWriter, r *http.Request) {
	if s.webhookSecret == "" {
		serveAPIErr(w, fmt.Errorf("server running without webhook secret"), http.StatusInternalServerError, "")
		return
	}

	ctx := r.Context()
	vars := mux.Vars(r)
	token := vars["token"]
	if token != s.webhookSecret {
		serveAPIErr(w, fmt.Errorf("invalid webhook secret"), http.StatusForbidden, "")
		return
	}

	var reqBody RandomPullRequest

	var b bytes.Buffer
	wrappedReader := io.TeeReader(r.Body, &b)
	err := json.NewDecoder(wrappedReader).Decode(&reqBody)
	if err != nil {
		slog.Error("error parsing payload", "payload", b.String())
		serveAPIErr(w, err, http.StatusBadRequest, "could not parse")
		return
	}
	defer r.Body.Close()

	slog.Info("RandomPull", "payload", reqBody)

	user, err := s.db.GetUser(ctx, db.GetUserOptions{
		TwitchID: reqBody.TwitchID,
	})
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {

			twusers, nerr := s.twitchClient.GetUsersByID(ctx, reqBody.TwitchID)
			if nerr != nil {
				serveAPIErr(w, nerr, http.StatusInternalServerError, "unable to get user from twitch")
				return
			}

			if len(twusers) < 1 {
				serveAPIErr(w, fmt.Errorf("wtf"), http.StatusInternalServerError, "unable to get user from twitch")
				return
			}
			twuser := twusers[0]

			user, nerr = s.db.CreateUser(ctx, db.User{
				Users: model.Users{
					ID:       s.idGenerator.Generate().Int64(),
					TwitchID: &twuser.ID,
					Name:     twuser.DisplayName,
				},
			})
			err = nerr
		}
		if err != nil {
			serveAPIErr(w, err, http.StatusInternalServerError, "unexpected error")
			return
		}
	}

	collectable, err := s.getRandomCollectable(ctx)
	if err != nil {
		serveAPIErr(w, err, http.StatusInternalServerError, "unexpected error")
		return
	}

	verified := (rand.Intn(100) == 0)

	tags, err := json.Marshal(map[string]bool{
		"subscriber": reqBody.Subscriber,
		"verified":   verified,
	})
	if err != nil {
		serveAPIErr(w, err, http.StatusInternalServerError, "unexpected error")
		return
	}
	tagString := string(tags)

	var issued *db.CollectableInstance
	if !reqBody.DryRun {
		issued, err = s.db.CreateCollectableInstance(ctx, model.CollectableInstances{
			ID:            s.idGenerator.Generate().Int64(),
			CollectableID: collectable.ID,
			OwnerID:       user.ID,
			EditionID:     1,
			CreatedAt:     time.Now().UTC(),
			Tags:          &tagString,
		})
		if err != nil {
			serveAPIErr(w, err, http.StatusInternalServerError, "unexpected error")
			return
		}
	} else {
		issued = &db.CollectableInstance{
			CollectableInstances: model.CollectableInstances{
				ID:            s.idGenerator.Generate().Int64(),
				CollectableID: collectable.ID,
				OwnerID:       user.ID,
				EditionID:     1,
				CreatedAt:     time.Now().UTC(),
				Tags:          &tagString,
			},
			Collectable: collectable,
			Owner:       &user.Users,
		}
	}

	c := IssuedCollectableFromCollectableInstance(issued)

	serveAPIPayload(
		w,
		&c,
	)
}

func (s *Server) getRandomCollectable(ctx context.Context) (*db.Collectable, error) {
	weights, err := s.db.GetWeights(ctx, 1)
	if err != nil {
		return nil, err
	}

	// Roll to Pick Rarity
	var sum int64 = 0
	for _, w := range weights {
		sum += int64(w.Weight)
	}

	// Roll to Pick Knife
	rarityRoll := rand.Int63n(sum)
	rarity := ""
	var acc int64 = 0
	for _, w := range weights {
		acc += int64(w.Weight)
		if rarityRoll < acc {
			rarity = w.Rarity
			break
		}
	}

	if rarity == "" {
		return nil, fmt.Errorf("unable to pick a rarity: %d %+v", rarityRoll, weights)
	}

	c, err := s.db.GetCollectables(ctx, db.GetCollectablesOptions{
		Collection: 1,
		Rarity:     rarity,
	})
	if err != nil {
		return nil, err
	}

	// Give me a random knifetype
	hit := c[rand.Intn(len(c))]

	return hit, nil
}

func (s *Server) adminGetCollectable(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	u, err := s.getAuthUser(ctx, r)
	if err != nil {
		serveAPIErr(w, err, http.StatusForbidden, "could not identify user")
		return
	}

	if u.Admin == nil || !*u.Admin {
		serveAPIErr(w, errAdminOnly, http.StatusForbidden, "")
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		serveAPIErr(w, err, http.StatusBadRequest, "id is non numeric")
		return
	}

	c, err := s.db.GetCollectable(ctx, id, db.GetCollectableOptions{
		GetDeleted:    true,
		GetUnapproved: true,
	})
	if err != nil {
		serveAPIErr(w, err, http.StatusInternalServerError, "unable to get collectable")
		return
	}

	serveAPIPayload(
		w,
		&struct {
			Collectable AdminCollectable
		}{
			Collectable: AdminCollectableFromDBCollectable(c),
		},
	)
}

func (s *Server) adminApproveCollectable(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	u, err := s.getAuthUser(ctx, r)
	if err != nil {
		serveAPIErr(w, err, http.StatusForbidden, "could not identify user")
		return
	}

	if u.Admin == nil || !*u.Admin {
		serveAPIErr(w, errAdminOnly, http.StatusForbidden, "")
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		serveAPIErr(w, err, http.StatusBadRequest, "id is non numeric")
		return
	}

	c, err := s.db.ApproveCollectable(ctx, id, u.ID)
	if err != nil {
		serveAPIErr(w, err, http.StatusInternalServerError, "unable to get collectable")
		return
	}

	serveAPIPayload(
		w,
		&struct {
			Collectable AdminCollectable
		}{
			Collectable: AdminCollectableFromDBCollectable(c),
		},
	)
}
