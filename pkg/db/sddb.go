package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
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
  knife_ownership.was_subscriber,
  knife_ownership.is_verified,
  editions.name,
  knife_ownership.transacted_at,
  knives.deleted
FROM knife_ownership
INNER JOIN knives ON knife_ownership.knife_id = knives.id
LEFT JOIN users owner ON knife_ownership.user_id = owner.id
LEFT JOIN users author ON knives.author_id = author.id
JOIN editions ON knife_ownership.edition_id = editions.id
WHERE owner.id != 166
AND knives.deleted = false
AND knives.approved_by is not null
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
			&knife.Subscriber,
			&knife.Verified,
			&knife.Edition,
			&obtainedAt,
			&knife.Deleted,
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
  knife_ownership.was_subscriber,
  knife_ownership.is_verified,
  editions.name,
  knife_ownership.transacted_at,
  knives.deleted
FROM knife_ownership
JOIN knives ON knife_ownership.knife_id = knives.id
LEFT JOIN users owner ON knife_ownership.user_id = owner.id
LEFT JOIN users author ON knives.author_id = author.id
JOIN editions ON knife_ownership.edition_id = editions.id
WHERE knife_ownership.instance_id = ?
AND knives.approved_by is not null;
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
		&knife.Subscriber,
		&knife.Verified,
		&knife.Edition,
		&obtainedAt,
		&knife.Deleted,
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
  knife_ownership.was_subscriber,
  knife_ownership.is_verified,
  editions.name,
  knife_ownership.transacted_at,
  knives.deleted
FROM knife_ownership
JOIN knives ON knife_ownership.knife_id = knives.id
LEFT JOIN users owner ON knife_ownership.user_id = owner.id
LEFT JOIN users author ON knives.author_id = author.id
JOIN editions ON knife_ownership.edition_id = editions.id
WHERE owner.lookup_name = ?
AND knives.deleted = false
AND knives.approved_by is not null
ORDER BY knife_ownership.transacted_at DESC;
`

func (sd *SDDB) GetKnivesForUsername(ctx context.Context, username string) ([]*Knife, error) {
	query, err := sd.db.PrepareContext(ctx, getKnifeForUserQuery)
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
			&knife.Subscriber,
			&knife.Verified,
			&knife.Edition,
			&obtainedAt,
			&knife.Deleted,
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

var getUserIDQuery = `
SELECT
  id,
  twitch_name,
  lookup_name,
  admin,
  IFNULL(twitch_id, '') as twitch_id,
  created_at
FROM users
WHERE id = ?;
`

func (sd *SDDB) GetUserByID(ctx context.Context, id int) (*User, error) {
	query, err := sd.db.Prepare(getUserIDQuery)
	if err != nil {
		return nil, err
	}

	rows, err := query.QueryContext(ctx, id)
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
		&user.Admin,
		&user.TwitchID,
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

var getUserTwitchIDQuery = `
SELECT
  id,
  twitch_name,
  lookup_name,
  IFNULL(twitch_id, '') as twitch_id,
  created_at
FROM users
WHERE twitch_id = ?;
`

func (sd *SDDB) GetUserByTwitchID(ctx context.Context, id string) (*User, error) {
	query, err := sd.db.Prepare(getUserTwitchIDQuery)
	if err != nil {
		return nil, err
	}

	rows, err := query.QueryContext(ctx, id)
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
		&user.TwitchID,
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

var getUserUsernameQuery = `
SELECT
  id,
  twitch_name,
  lookup_name,
  IFNULL(twitch_id, '') as twitch_id,
  created_at
FROM users
WHERE lookup_name = ?;
`

func (sd *SDDB) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	query, err := sd.db.Prepare(getUserUsernameQuery)
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
		&user.TwitchID,
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

var getUsersSearchQuery = `
SELECT
  id,
  twitch_name,
  lookup_name,
  IFNULL(twitch_id, '') as twitch_id,
  created_at
FROM users
WHERE lookup_name LIKE CONCAT('%%', ?, '%%')
LIMIT 10;
`

func (sd *SDDB) GetUsers(ctx context.Context, substr string) ([]*User, error) {
	if substr == "" {
		return nil, fmt.Errorf("search string required")
	}

	query, err := sd.db.Prepare(getUsersSearchQuery)
	if err != nil {
		return nil, err
	}

	rows, err := query.QueryContext(ctx, substr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []*User{}
	for rows.Next() {
		var createdAt string

		var user User
		err = rows.Scan(
			&user.ID,
			&user.Name,
			&user.LookupName,
			&user.TwitchID,
			&createdAt,
		)
		if err != nil {
			return nil, err
		}

		user.CreatedAt, err = parseTimestamp(createdAt)
		if err != nil {
			return nil, err
		}

		users = append(users, &user)
	}

	return users, nil
}

var createUserQuery = `INSERT INTO users (twitch_name, lookup_name, twitch_id) VALUES (?, ?, ?);`

var createUserByNameQuery = `INSERT INTO users (twitch_name, lookup_name) VALUES (?, ?);`

func (sd *SDDB) CreateUser(ctx context.Context, user *User) (*User, error) {
	if user == nil {
		return nil, fmt.Errorf("user cannot be nil")
	}

	if user.Name == "" {
		return nil, fmt.Errorf("username must be specified")
	}

	lookupName := user.LookupName
	if lookupName == "" {
		lookupName = strings.ToLower(user.Name)
	}

	var query *sql.Stmt
	var err error
	var res sql.Result
	if user.TwitchID != "" {
		query, err = sd.db.Prepare(createUserQuery)
		if err != nil {
			return nil, err
		}

		res, err = query.ExecContext(ctx, user.Name, lookupName, user.TwitchID)
		if err != nil {
			return nil, err
		}
	} else {
		query, err = sd.db.Prepare(createUserByNameQuery)
		if err != nil {
			return nil, err
		}

		res, err = query.ExecContext(ctx, user.Name, lookupName)
		if err != nil {
			return nil, err
		}
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &User{
		ID:         int(id),
		Name:       user.Name,
		LookupName: lookupName,
		TwitchID:   user.TwitchID,
		Admin:      false,
		CreatedAt:  time.Now(),
	}, nil
}

var (
	getKnifeByName       = `SELECT id FROM knives WHERE name = ? AND deleted = false;`
	createKnifePullQuery = `INSERT INTO knife_ownership (user_id, knife_id, trans_type, was_subscriber, is_verified, edition_id) VALUES (?, ?, ?, ?, ?, ?);`
)

func (sd *SDDB) PullKnife(ctx context.Context, userID int, knifename string, subscriber bool, verified bool, edition_id int) (*Knife, error) {
	// Lookup knifeID by name
	knifeq, err := sd.db.Prepare(getKnifeByName)
	if err != nil {
		return nil, err
	}

	rows, err := knifeq.Query(knifename)
	if err != nil {
		return nil, ErrNotFound
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, ErrNotFound
	}

	var knifeID int

	err = rows.Scan(&knifeID)
	if err != nil {
		return nil, err
	}

	// Create the pull
	createQ, err := sd.db.PrepareContext(ctx, createKnifePullQuery)
	if err != nil {
		return nil, err
	}

	res, err := createQ.ExecContext(ctx, userID, knifeID, "pull", subscriber, verified, edition_id)
	if err != nil {
		return nil, err
	}

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
  knives.deleted,
  knives.approved_by,
  knives.approved_at
FROM knives
LEFT JOIN users author ON knives.author_id = author.id
WHERE knives.approved_by is not null
%s
ORDER BY knives.id ASC;
`

func (sd *SDDB) GetCollection(ctx context.Context, getDeleted bool) ([]*KnifeType, error) {
	optionalWhere := "AND knives.deleted = false"
	if getDeleted {
		optionalWhere = ""
	}

	q, err := sd.db.PrepareContext(ctx, fmt.Sprintf(getCatalogQuery, optionalWhere))
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

		var approvedBy *int
		var approvedAt *string

		err = rows.Scan(
			&k.ID,
			&k.Name,
			&k.Author,
			&k.AuthorID,
			&k.Rarity,
			&k.ImageName,
			&k.Deleted,
			&approvedBy,
			&approvedAt,
		)
		if err != nil {
			slog.Warn("deserializing GetCollection", "err", err)
			continue
		}

		k.Approved = (approvedBy != nil)
		if approvedAt != nil {
			k.ApprovedAt, err = parseTimestamp(*approvedAt)
			if err != nil {
				return nil, err
			}
		}

		knives = append(knives, &k)
	}

	return knives, nil
}

var getPendingKnifeQuery = `
SELECT
  knives.id,
  knives.name,
  author.twitch_name,
  author.id,
  knives.rarity,
  knives.image_name,
  knives.deleted,
  knives.approved_by,
  knives.approved_at
FROM knives
LEFT JOIN users author ON knives.author_id = author.id
WHERE knives.approved_by is null
AND knives.deleted = false
ORDER BY knives.created_at ASC;
`

func (sd *SDDB) GetPendingKnives(ctx context.Context) ([]*KnifeType, error) {
	q, err := sd.db.PrepareContext(ctx, getPendingKnifeQuery)
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

		var approvedBy *int
		var approvedAt *string

		err = rows.Scan(
			&k.ID,
			&k.Name,
			&k.Author,
			&k.AuthorID,
			&k.Rarity,
			&k.ImageName,
			&k.Deleted,
			&approvedBy,
			&approvedAt,
		)
		if err != nil {
			slog.Warn("deserializing GetPendingKnives", "err", err)
			continue
		}

		k.Approved = (approvedBy != nil)
		if approvedAt != nil {
			k.ApprovedAt, err = parseTimestamp(*approvedAt)
			if err != nil {
				return nil, err
			}
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
  knives.deleted,
  knives.approved_by,
  knives.approved_at
FROM knives
LEFT JOIN users author ON knives.author_id = author.id
WHERE knives.id = ?
%s
LIMIT 1;
`

func (sd *SDDB) GetKnifeType(ctx context.Context, id int, getDeleted bool, getUnapproved bool) (*KnifeType, error) {
	optionalWhere := ""
	if !getDeleted {
		optionalWhere += "AND knives.deleted = false "
	}
	if !getUnapproved {
		optionalWhere += "AND knives.approved_by is not null "
	}

	q, err := sd.db.PrepareContext(ctx, fmt.Sprintf(getKnifeTypeQuery, optionalWhere))
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
	var approvedBy *int
	var approvedAt *string

	err = rows.Scan(
		&k.ID,
		&k.Name,
		&k.Author,
		&k.AuthorID,
		&k.Rarity,
		&k.ImageName,
		&k.Deleted,
		&approvedBy,
		&approvedAt,
	)
	if err != nil {
		return nil, err
	}

	k.Approved = (approvedBy != nil)

	if approvedAt != nil {
		k.ApprovedAt, err = parseTimestamp(*approvedAt)
		if err != nil {
			return nil, err
		}
	}

	return &k, nil
}

var getKnifeTypeRarityQuery = `
SELECT
  knives.id,
  knives.name,
  author.twitch_name,
  author.id,
  knives.rarity,
  knives.image_name,
  knives.deleted,
  knives.approved_by,
  knives.approved_at
FROM knives
LEFT JOIN users author ON knives.author_id = author.id
WHERE knives.rarity = ?
AND knives.deleted = false
AND knives.approved_by is not null;
`

func (sd *SDDB) GetKnifeTypesByRarity(ctx context.Context, rarity string) ([]*KnifeType, error) {
	q, err := sd.db.PrepareContext(ctx, getKnifeTypeRarityQuery)
	if err != nil {
		return nil, err
	}

	rows, err := q.QueryContext(ctx, rarity)
	if err != nil {
		return nil, err
	}

	res := []*KnifeType{}
	for rows.Next() {
		var k KnifeType
		var approvedBy *int
		var approvedAt *string

		err = rows.Scan(
			&k.ID,
			&k.Name,
			&k.Author,
			&k.AuthorID,
			&k.Rarity,
			&k.ImageName,
			&k.Deleted,
			&approvedBy,
			&approvedAt,
		)
		if err != nil {
			return nil, err
		}

		k.Approved = (approvedBy != nil)

		if approvedAt != nil {
			k.ApprovedAt, err = parseTimestamp(*approvedAt)
			if err != nil {
				return nil, err
			}
		}

		res = append(res, &k)
	}

	return res, nil
}

var createKnifeTypeQuery = `
INSERT INTO knives (name, author_id, rarity, image_name) VALUES (?, ?, ?, ?);
`

func (sd *SDDB) CreateKnifeType(ctx context.Context, knife *KnifeType) (*KnifeType, error) {
	q, err := sd.db.PrepareContext(ctx, createKnifeTypeQuery)
	if err != nil {
		return nil, err
	}

	res, err := q.ExecContext(ctx, knife.Name, knife.AuthorID, knife.Rarity, knife.ImageName)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	return sd.GetKnifeType(ctx, int(id), true, true)
}

var createEditionQuery = `
INSERT INTO editions (name) VALUES (?);
`

func (sd *SDDB) CreateEdition(ctx context.Context, edition *Edition) (*Edition, error) {
	q, err := sd.db.PrepareContext(ctx, createEditionQuery)
	if err != nil {
		return nil, err
	}

	res, err := q.ExecContext(ctx, edition.Name)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	return sd.GetEdition(ctx, int(id))
}

var getEditionQuery = `
SELECT 
  id,
  name,
  updated_at
FROM editions
WHERE id = ?
LIMIT 1;
`

// GetEdition returns the edition with the given ID.
func (sd *SDDB) GetEdition(ctx context.Context, id int) (*Edition, error) {
	q, err := sd.db.PrepareContext(ctx, getEditionQuery)
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

	var edition Edition
	var timeString string
	err = rows.Scan(
		&edition.ID,
		&edition.Name,
		&timeString,
	)
	if err != nil {
		return nil, err
	}

	edition.UpdatedAt, err = parseTimestamp(timeString)
	if err != nil {
		return nil, err
	}

	return &edition, nil
}

var getEditionsQuery = `
SELECT 
  id,
  name,
  updated_at
FROM editions
ORDER BY id ASC;
`

func (sd *SDDB) GetEditions(ctx context.Context) ([]*Edition, error) {
	q, err := sd.db.PrepareContext(ctx, getEditionsQuery)
	if err != nil {
		return nil, err
	}

	rows, err := q.QueryContext(ctx)
	if err != nil {
		return nil, err
	}

	res := []*Edition{}

	for rows.Next() {
		var edition Edition
		var timeString string
		err := rows.Scan(
			&edition.ID,
			&edition.Name,
			&timeString,
		)
		if err != nil {
			return nil, err
		}

		edition.UpdatedAt, err = parseTimestamp(timeString)
		if err != nil {
			return nil, err
		}

		res = append(res, &edition)
	}

	return res, nil
}

var deleteKnifeByID = `
UPDATE knives SET deleted = true WHERE id = ?;
`

func (sd *SDDB) DeleteKnifeType(ctx context.Context, knife *KnifeType) error {
	q, err := sd.db.PrepareContext(ctx, deleteKnifeByID)
	if err != nil {
		return err
	}

	if knife.ID == 0 {
		return fmt.Errorf("knife must be specified")
	}

	_, err = q.ExecContext(ctx, knife.ID)
	if err != nil {
		return err
	}

	return nil
}

var updateKnifeTypeQuery = `
UPDATE knives SET name = ?, author_id = ?, rarity = ?, image_name = ? WHERE id = ?;
`

func (sd *SDDB) UpdateKnifeType(ctx context.Context, knife *KnifeType) (*KnifeType, error) {
	q, err := sd.db.PrepareContext(ctx, updateKnifeTypeQuery)
	if err != nil {
		return nil, err
	}

	_, err = q.ExecContext(ctx, knife.Name, knife.AuthorID, knife.Rarity, knife.ImageName, knife.ID)
	if err != nil {
		return nil, err
	}

	return sd.GetKnifeType(ctx, knife.ID, true, true)
}

var equipKnifeQuery = `
INSERT INTO equipped (user_id, instance_id) 
VALUES (?, ?)
ON DUPLICATE KEY UPDATE
  user_id = ?, instance_id = ?;
`

func (sd *SDDB) EquipKnifeForUser(ctx context.Context, userID int, knifeID int) error {
	q, err := sd.db.PrepareContext(ctx, equipKnifeQuery)
	if err != nil {
		return err
	}

	_, err = q.ExecContext(ctx, userID, knifeID, userID, knifeID)
	if err != nil {
		return err
	}

	return nil
}

var getEquippedKnifeQuery = `
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
  knife_ownership.was_subscriber,
  knife_ownership.is_verified,
  editions.name,
  knife_ownership.transacted_at,
  knives.deleted
FROM knife_ownership
JOIN knives ON knife_ownership.knife_id = knives.id
LEFT JOIN users owner ON knife_ownership.user_id = owner.id
LEFT JOIN users author ON knives.author_id = author.id
JOIN editions ON knife_ownership.edition_id = editions.id
WHERE knife_ownership.instance_id = (SELECT instance_id FROM equipped WHERE user_id = ?)
AND knives.deleted = false;
`

func (sd *SDDB) GetEquippedKnifeForUser(ctx context.Context, userID int) (*Knife, error) {
	q, err := sd.db.PrepareContext(ctx, getEquippedKnifeQuery)
	if err != nil {
		return nil, err
	}

	rows, err := q.QueryContext(ctx, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
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
		&knife.Subscriber,
		&knife.Verified,
		&knife.Edition,
		&obtainedAt,
		&knife.Deleted,
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

var insertImageUpload = `
INSERT INTO image_uploads (user_id, image_id, path, uploadname)
VALUES (?, ?, ?, ?);
`

func (sd *SDDB) CreateImageUpload(ctx context.Context, id int64, authorID int, path string, uploadname string) error {
	query, err := sd.db.Prepare(insertImageUpload)
	if err != nil {
		return err
	}

	_, err = query.ExecContext(ctx, authorID, id, path, uploadname)
	if err != nil {
		return err
	}
	return nil
}

func (sd *SDDB) IssueCollectable(ctx context.Context, collectableID int, userID int, subscriber bool, verified bool, editionID int, source string) (*Knife, error) {
	createQ, err := sd.db.PrepareContext(ctx, createKnifePullQuery)
	if err != nil {
		return nil, err
	}

	res, err := createQ.ExecContext(ctx, userID, collectableID, source, subscriber, verified, editionID)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	return sd.GetKnife(ctx, int(id))
}

var getWeightQuery = `
SELECT
  community_id,
  rarity,
  weight,
  updated_at
FROM pullconfig
WHERE community_id = 1;
`

func (sd *SDDB) GetWeights(ctx context.Context) ([]*PullWeight, error) {
	q, err := sd.db.PrepareContext(ctx, getWeightQuery)
	if err != nil {
		return nil, err
	}

	rows, err := q.QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := []*PullWeight{}

	for rows.Next() {
		var pw PullWeight

		var updatedAt string

		err = rows.Scan(
			&pw.CommunityID,
			&pw.Rarity,
			&pw.Weight,
			&updatedAt,
		)
		if err != nil {
			return nil, err
		}

		pw.UpdatedAt, err = parseTimestamp(updatedAt)
		if err != nil {
			return nil, err
		}

		res = append(res, &pw)
	}

	return res, nil
}

func (sd *SDDB) SetWeights(ctx context.Context, weights []*PullWeight) ([]*PullWeight, error) {
	return nil, fmt.Errorf("NOT IMPLEMENTED")
}

var approveKnifeQuery = `
UPDATE knives SET
approved_by = ?,
approved_at = CURRENT_TIMESTAMP()
WHERE id = ?;
`

func (sd *SDDB) ApproveKnifeType(ctx context.Context, id int, userID int) (*KnifeType, error) {
	createQ, err := sd.db.PrepareContext(ctx, approveKnifeQuery)
	if err != nil {
		return nil, err
	}

	_, err = createQ.ExecContext(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	return sd.GetKnifeType(ctx, id, true, true)
}

var getCombatReportQuery = `
SELECT
  id,
  participants,
  outcomes,
  knives,
  created_at
FROM fights
WHERE id = ?;
`

func (sd *SDDB) GetCombatReport(ctx context.Context, id int64) (*CombatReport, error) {
	q, err := sd.db.PrepareContext(ctx, getCombatReportQuery)
	if err != nil {
		return nil, err
	}

	var report CombatReport
	var createdAt string
	var participants string
	var outcomes string
	var knives string

	err = q.QueryRowContext(ctx, id).Scan(
		&report.ID,
		&participants,
		&outcomes,
		&knives,
		&createdAt,
	)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(participants), &report.Participants)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(outcomes), &report.Outcomes)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(knives), &report.Knives)
	if err != nil {
		return nil, err
	}

	report.CreatedAt, err = parseTimestamp(createdAt)
	if err != nil {
		return nil, err
	}

	return &report, nil
}

var insertCombatReportQuery = `
INSERT INTO fights (id, participants, outcomes, knives) VALUES (?, ?, ?, ?);
`

func (sd *SDDB) CreateCombatReport(ctx context.Context, report *CombatReport) (*CombatReport, error) {
	q, err := sd.db.PrepareContext(ctx, insertCombatReportQuery)
	if err != nil {
		return nil, err
	}

	participants, err := json.Marshal(report.Participants)
	if err != nil {
		return nil, fmt.Errorf("encoding participants: %w", err)
	}
	outcomes, err := json.Marshal(report.Outcomes)
	if err != nil {
		return nil, fmt.Errorf("encoding outcomes: %w", err)
	}
	knives, err := json.Marshal(report.Knives)
	if err != nil {
		return nil, fmt.Errorf("encoding outcomes: %w", err)
	}

	_, err = q.ExecContext(ctx, report.ID, participants, outcomes, knives)
	if err != nil {
		return nil, err
	}

	return sd.GetCombatReport(ctx, report.ID)
}
