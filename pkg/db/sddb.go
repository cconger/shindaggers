package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var ErrNotFound = errors.New("not found")

type SDDB struct {
	db *sql.DB
}

const mysqlTimestampFmt = "2006-01-02 15:04:05"

func parseTimestamp(ts string) (time.Time, error) {
	return time.Parse(mysqlTimestampFmt, ts)
}

func NewSDDB(connectionString string) (KnifeDB, error) {
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &SDDB{
		db: db,
	}, nil
}

func (sd *SDDB) Close(ctx context.Context) error {
	return sd.db.Close()
}

var getLatestPullsQuery = `
SELECT
  knives.id,
  knife_ownership.instance_id,
  knives.name,
  author.twitch_name,
  author.id,
  owner.twitch_name,
  owner.id,
  knives.rarity,
  knives.image_name,
  editions.name,
  knife_ownership.transacted_at
FROM knife_ownership
JOIN knives ON knife_ownership.knife_id = knives.id
LEFT JOIN users owner ON knife_ownership.user_id = owner.id
LEFT JOIN users author ON knives.author_id = author.id
JOIN editions ON knives.edition_id = editions.id
WHERE owner.id != 166
ORDER BY knife_ownership.transacted_at DESC
LIMIT 15;
`

func (sd *SDDB) GetLatestPulls(ctx context.Context) ([]*Knife, error) {
	query, err := sd.db.Prepare(getLatestPullsQuery)
	if err != nil {
		return nil, err
	}

	rows, err := query.QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	knives := []*Knife{}
	for rows.Next() {
		var obtainedAt string

		var knife Knife
		err = rows.Scan(
			&knife.ID,
			&knife.InstanceID,
			&knife.Name,
			&knife.Author,
			&knife.AuthorID,
			&knife.Owner,
			&knife.OwnerID,
			&knife.Rarity,
			&knife.ImageName,
			&knife.Edition,
			&obtainedAt,
		)
		if err != nil {
			return nil, err
		}

		knife.ObtainedAt, err = parseTimestamp(obtainedAt)
		if err != nil {
			return nil, err
		}

		knives = append(knives, &knife)
	}

	return knives, nil
}

var getKnifeQuery = `
SELECT
  knives.id,
  knife_ownership.instance_id,
  knives.name,
  author.twitch_name,
  author.id,
  owner.twitch_name,
  owner.id,
  knives.rarity,
  knives.image_name,
  editions.name,
  knife_ownership.transacted_at
FROM knife_ownership
JOIN knives ON knife_ownership.knife_id = knives.id
LEFT JOIN users owner ON knife_ownership.user_id = owner.id
LEFT JOIN users author ON knives.author_id = author.id
JOIN editions ON knives.edition_id = editions.id
WHERE knife_ownership.instance_id = ?;
`

func (sd *SDDB) GetKnife(ctx context.Context, knifeID int) (*Knife, error) {
	query, err := sd.db.Prepare(getKnifeQuery)
	if err != nil {
		return nil, err
	}

	rows, err := query.QueryContext(ctx, knifeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, ErrNotFound
	}

	var obtainedAt string

	var knife Knife
	err = rows.Scan(
		&knife.ID,
		&knife.InstanceID,
		&knife.Name,
		&knife.Author,
		&knife.AuthorID,
		&knife.Owner,
		&knife.OwnerID,
		&knife.Rarity,
		&knife.ImageName,
		&knife.Edition,
		&obtainedAt,
	)
	if err != nil {
		return nil, err
	}

	knife.ObtainedAt, err = parseTimestamp(obtainedAt)
	if err != nil {
		return nil, err
	}

	return &knife, nil
}

var getKnifeForUserQuery = `
SELECT
  knives.id,
  knife_ownership.instance_id,
  knives.name,
  author.twitch_name,
  author.id,
  owner.twitch_name,
  owner.id,
  knives.rarity,
  knives.image_name,
  editions.name,
  knife_ownership.transacted_at
FROM knife_ownership
JOIN knives ON knife_ownership.knife_id = knives.id
LEFT JOIN users owner ON knife_ownership.user_id = owner.id
LEFT JOIN users author ON knives.author_id = author.id
JOIN editions ON knives.edition_id = editions.id
WHERE owner.lookup_name = ?
ORDER BY knife_ownership.transacted_at DESC;
`

func (sd *SDDB) GetKnivesForUsername(ctx context.Context, username string) ([]*Knife, error) {
	query, err := sd.db.Prepare(getKnifeForUserQuery)
	if err != nil {
		return nil, err
	}

	rows, err := query.QueryContext(ctx, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	knives := []*Knife{}

	for rows.Next() {
		var obtainedAt string

		var knife Knife
		err = rows.Scan(
			&knife.ID,
			&knife.InstanceID,
			&knife.Name,
			&knife.Author,
			&knife.AuthorID,
			&knife.Owner,
			&knife.OwnerID,
			&knife.Rarity,
			&knife.ImageName,
			&knife.Edition,
			&obtainedAt,
		)
		if err != nil {
			return nil, err
		}

		knife.ObtainedAt, err = parseTimestamp(obtainedAt)
		if err != nil {
			return nil, err
		}

		knives = append(knives, &knife)
	}

	return knives, nil
}

var getUserQuery = `
SELECT
  id,
  twitch_name,
  lookup_name,
  created_at
FROM users
WHERE lookup_name = ?;
`

func (sd *SDDB) GetUser(ctx context.Context, username string) (*User, error) {
	query, err := sd.db.Prepare(getUserQuery)
	if err != nil {
		return nil, err
	}

	rows, err := query.QueryContext(ctx, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, ErrNotFound
	}

	var createdAt string

	var user User
	err = rows.Scan(
		&user.ID,
		&user.Name,
		&user.LookupName,
		&createdAt,
	)
	if err != nil {
		return nil, err
	}

	user.CreatedAt, err = parseTimestamp(createdAt)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

var createUserQuery = `INSERT INTO users (twitch_name, lookup_name, twitch_id) VALUES (?, ?, ?);`

func (sd *SDDB) CreateUser(ctx context.Context, user *User) (*User, error) {
	if user == nil {
		return nil, fmt.Errorf("user cannot be nil")
	}
	query, err := sd.db.Prepare(createUserQuery)
	if err != nil {
		return nil, err
	}

	if user.Name == "" {
		return nil, fmt.Errorf("username must be specified")
	}

	lookupName := user.LookupName
	if lookupName == "" {
		lookupName = strings.ToLower(user.Name)
	}

	_, err = query.ExecContext(ctx, user.Name, lookupName, user.TwitchID)
	if err != nil {
		return nil, err
	}

	return sd.GetUser(ctx, lookupName)
}

var (
	getKnifeByName       = `SELECT id FROM knives WHERE name = ?;`
	createKnifePullQuery = `INSERT INTO knife_ownership (user_id, knife_id, trans_type) VALUES (?, ?, ?);`
)

func (sd *SDDB) PullKnife(ctx context.Context, username string, knifename string) (*Knife, error) {
	// TODO: Transactions

	// Lookup knifeID by name
	knifeq, err := sd.db.Prepare(getKnifeByName)
	if err != nil {
		return nil, err
	}

	rows, err := knifeq.Query(knifename)
	if err != nil {
		return nil, ErrNotFound
	}

	if !rows.Next() {
		return nil, ErrNotFound
	}

	var knifeID int

	err = rows.Scan(&knifeID)
	if err != nil {
		return nil, err
	}

	// Lookup user by name (create if missing)
	user, err := sd.GetUser(ctx, username)
	if err != nil {
		if err == ErrNotFound {
			user, err = sd.CreateUser(ctx, &User{
				Name: username,
			})
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	// Create the pull
	createQ, err := sd.db.PrepareContext(ctx, createKnifePullQuery)
	if err != nil {
		return nil, err
	}

	res, err := createQ.ExecContext(ctx, user.ID, knifeID, "pull")
	if err != nil {
		return nil, err
	}

	log.Printf(
		"Created pull for %d of knife %d based on inputs %s and %s",
		user.ID,
		knifeID,
		username,
		knifename,
	)

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	return sd.GetKnife(ctx, int(id))
}

var queryUserAuthByToken = `
SELECT
  user_id,
  token,
  access_token,
  refresh_token,
  expires_at,
  updated_at
FROM user_auth
WHERE token = ?;
`

func (sd *SDDB) GetAuth(ctx context.Context, token []byte) (*UserAuth, error) {
	getTokenQ, err := sd.db.PrepareContext(ctx, queryUserAuthByToken)
	if err != nil {
		return nil, err
	}

	rows, err := getTokenQ.QueryContext(ctx, token)
	if err != nil {
		return nil, err
	}

	if !rows.Next() {
		return nil, ErrNotFound
	}

	var expiresAtStr string
	var updatedAtStr string
	var userAuth UserAuth
	err = rows.Scan(
		&userAuth.UserID,
		&userAuth.Token,
		&userAuth.AccessToken,
		&userAuth.RefreshToken,
		&expiresAtStr,
		&updatedAtStr,
	)
	if err != nil {
		return nil, err
	}

	userAuth.ExpiresAt, err = parseTimestamp(expiresAtStr)
	if err != nil {
		return nil, err
	}
	userAuth.UpdatedAt, err = parseTimestamp(updatedAtStr)
	if err != nil {
		return nil, err
	}

	return &userAuth, nil
}

var saveAuthQuery = `
INSERT INTO user_auth (user_id, token, access_token, refresh_token, expires_at) VALUES (?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE token = VALUES(token), access_token = VALUES(access_token), refresh_token = VALUES(refresh_token), expires_at = VALUES(expires_at);
`

func (sd *SDDB) SaveAuth(ctx context.Context, auth *UserAuth) (*UserAuth, error) {
	// TODO: Validate required fields on auth

	query, err := sd.db.PrepareContext(ctx, saveAuthQuery)
	if err != nil {
		return nil, err
	}

	_, err = query.ExecContext(
		ctx,
		auth.UserID,
		auth.Token,
		auth.AccessToken,
		auth.RefreshToken,
		auth.ExpiresAt,
	)
	if err != nil {
		return nil, err
	}

	// TODO: this should actually be the query... but w/e
	return auth, nil
}

var getCatalogQuery = `
SELECT
  knives.id,
  knives.name,
  author.twitch_name,
  author.id,
  knives.rarity,
  knives.image_name,
  editions.name
FROM knives
LEFT JOIN users author ON knives.author_id = author.id
JOIN editions ON knives.edition_id = editions.id
ORDER BY knives.edition_id, knives.id ASC;
`

func (sd *SDDB) GetCollection(ctx context.Context) ([]*KnifeType, error) {
	q, err := sd.db.PrepareContext(ctx, getCatalogQuery)
	if err != nil {
		return nil, err
	}

	rows, err := q.QueryContext(ctx)
	if err != nil {
		return nil, err
	}

	knives := []*KnifeType{}
	for rows.Next() {
		var k KnifeType

		err = rows.Scan(
			&k.ID,
			&k.Name,
			&k.Author,
			&k.AuthorID,
			&k.Rarity,
			&k.ImageName,
			&k.Edition,
		)
		if err != nil {
			log.Printf("Error: scan GetCollection: %s", err)
			continue
		}

		knives = append(knives, &k)
	}

	return knives, nil
}

var getKnifeTypeQuery = `
SELECT
  knives.id,
  knives.name,
  author.twitch_name,
  author.id,
  knives.rarity,
  knives.image_name,
  editions.name
FROM knives
LEFT JOIN users author ON knives.author_id = author.id
JOIN editions ON knives.edition_id = editions.id
WHERE knives.id = ?
LIMIT 1;
`

func (sd *SDDB) GetKnifeType(ctx context.Context, id int) (*KnifeType, error) {
	q, err := sd.db.PrepareContext(ctx, getKnifeTypeQuery)
	if err != nil {
		return nil, err
	}

	rows, err := q.QueryContext(ctx, id)
	if err != nil {
		return nil, err
	}

	if !rows.Next() {
		return nil, ErrNotFound
	}

	var k KnifeType

	err = rows.Scan(
		&k.ID,
		&k.Name,
		&k.Author,
		&k.AuthorID,
		&k.Rarity,
		&k.ImageName,
		&k.Edition,
	)
	if err != nil {
		log.Printf("Error: scan GetCollection: %s", err)
		return nil, err
	}

	return &k, nil
}
