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

func (sd *SDDB) GetKnife(ctx context.Context, knifeID int64) (*Knife, error) {
	k, err := sd.GetKnives(ctx, knifeID)
	if err != nil {
		return nil, err
	}
	if len(k) == 0 {
		return nil, ErrNotFound
	}

	return k[0], nil
}

var getKnifesQuery = `
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
WHERE knife_ownership.instance_id IN (%s)
AND knives.approved_by is not null;
`

func (sd *SDDB) GetKnives(ctx context.Context, knifeID ...int64) ([]*Knife, error) {
	if len(knifeID) == 0 {
		return nil, nil
	}

	query, err := sd.db.Prepare(fmt.Sprintf(getKnifesQuery, strings.Repeat("?,", len(knifeID)-1)+"?"))
	if err != nil {
		return nil, err
	}

	idList := []interface{}{}
	for _, id := range knifeID {
		idList = append(idList, id)
	}

	rows, err := query.QueryContext(ctx, idList...)
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
WHERE owner.id = ?
AND knives.deleted = false
AND knives.approved_by is not null
ORDER BY knife_ownership.transacted_at DESC;
`

func (sd *SDDB) GetKnivesForUser(ctx context.Context, userID int64) ([]*Knife, error) {
	query, err := sd.db.PrepareContext(ctx, getKnifeForUserQuery)
	if err != nil {
		return nil, err
	}

	rows, err := query.QueryContext(ctx, userID)
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

func (sd *SDDB) GetUserByID(ctx context.Context, id int64) (*User, error) {
	us, err := sd.GetUsersByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if len(us) == 0 {
		return nil, ErrNotFound
	}

	return us[0], nil
}

var getMultipleUserIDQuery = `
SELECT
  id,
  twitch_name,
  lookup_name,
  admin,
  IFNULL(twitch_id, '') as twitch_id,
  created_at
FROM users
WHERE id IN (%s);
`

func (sd *SDDB) GetUsersByID(ctx context.Context, ids ...int64) ([]*User, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	query, err := sd.db.Prepare(fmt.Sprintf(getMultipleUserIDQuery, strings.Repeat("?,", len(ids)-1)+"?"))
	if err != nil {
		return nil, err
	}

	idList := []interface{}{}
	for _, id := range ids {
		idList = append(idList, id)
	}

	rows, err := query.QueryContext(ctx, idList...)
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

		users = append(users, &user)
	}

	return users, nil
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

var createUserQuery = `INSERT INTO users (id, twitch_name, lookup_name, twitch_id) VALUES (?, ?, ?, ?);`

func (sd *SDDB) CreateUser(ctx context.Context, user *User) (*User, error) {
	if user == nil {
		return nil, fmt.Errorf("user cannot be nil")
	}

	if user.ID == 0 {
		return nil, fmt.Errorf("user ID must be specified")
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
	query, err = sd.db.Prepare(createUserQuery)
	if err != nil {
		return nil, err
	}

	_, err = query.ExecContext(ctx, user.ID, user.Name, lookupName, user.TwitchID)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:         user.ID,
		Name:       user.Name,
		LookupName: lookupName,
		TwitchID:   user.TwitchID,
		Admin:      false,
		CreatedAt:  time.Now(),
	}, nil
}

var updateUserQuery = `UPDATE users SET twitch_name = ?, lookup_name = ?, twitch_id = ? WHERE id = ?;`

func (sd *SDDB) UpdateUser(ctx context.Context, user *User) (*User, error) {
	if user == nil {
		return nil, fmt.Errorf("user cannot be nil")
	}

	if user.ID == 0 {
		return nil, fmt.Errorf("user ID must be specified")
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
	query, err = sd.db.Prepare(updateUserQuery)
	if err != nil {
		return nil, err
	}

	_, err = query.ExecContext(ctx, user.Name, lookupName, user.TwitchID, user.ID)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:         user.ID,
		Name:       user.Name,
		LookupName: lookupName,
		TwitchID:   user.TwitchID,
		Admin:      false,
		CreatedAt:  time.Now(),
	}, nil
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

		var approvedBy *int64
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

func (sd *SDDB) GetKnifeType(ctx context.Context, id int64, getDeleted bool, getUnapproved bool) (*KnifeType, error) {
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
	var approvedBy *int64
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

var getKnifeTypeNameQuery = `
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
WHERE knives.name = ?
LIMIT 1;
`

func (sd *SDDB) GetKnifeTypeByName(ctx context.Context, name string) (*KnifeType, error) {
	q, err := sd.db.PrepareContext(ctx, getKnifeTypeQuery)
	if err != nil {
		return nil, err
	}

	rows, err := q.QueryContext(ctx, name)
	if err != nil {
		return nil, err
	}

	if !rows.Next() {
		return nil, ErrNotFound
	}

	var k KnifeType
	var approvedBy *int64
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
		var approvedBy *int64
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
INSERT INTO knives (id, name, author_id, rarity, image_name) VALUES (?, ?, ?, ?, ?);
`

func (sd *SDDB) CreateKnifeType(ctx context.Context, knife *KnifeType) (*KnifeType, error) {
	q, err := sd.db.PrepareContext(ctx, createKnifeTypeQuery)
	if err != nil {
		return nil, err
	}

	if knife.ID == 0 {
		return nil, errors.New("knife id must be set")
	}

	_, err = q.ExecContext(ctx, knife.ID, knife.Name, knife.AuthorID, knife.Rarity, knife.ImageName)
	if err != nil {
		return nil, err
	}

	return sd.GetKnifeType(ctx, knife.ID, true, true)
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

func (sd *SDDB) EquipKnifeForUser(ctx context.Context, userID int64, knifeID int64) error {
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

func (sd *SDDB) GetEquippedKnifeForUser(ctx context.Context, userID int64) (*Knife, error) {
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

func (sd *SDDB) CreateImageUpload(ctx context.Context, id int64, authorID int64, path string, uploadname string) error {
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

var issueCollectableQuery = `
INSERT INTO knife_ownership (
  instance_id,
  user_id,
  knife_id,
  trans_type,
  was_subscriber,
  is_verified,
  edition_id
) VALUES (?, ?, ?, ?, ?, ?, ?);
`

func (sd *SDDB) IssueCollectable(ctx context.Context, knife *Knife, source string) (*Knife, error) {
	createQ, err := sd.db.PrepareContext(ctx, issueCollectableQuery)
	if err != nil {
		return nil, err
	}

	_, err = createQ.ExecContext(ctx,
		knife.InstanceID,
		knife.OwnerID,
		knife.ID,
		source,
		knife.Subscriber,
		knife.Verified,
		1,
	)
	if err != nil {
		return nil, err
	}

	return sd.GetKnife(ctx, knife.InstanceID)
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

func (sd *SDDB) ApproveKnifeType(ctx context.Context, id int64, userID int64) (*KnifeType, error) {
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
  event,
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
		&report.Event,
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

var (
	insertCombatReportQuery = `
INSERT INTO fights (id, participants, outcomes, knives, event) VALUES (?, ?, ?, ?, ?);
`
	insertCombatOutcomeQuery = `
INSERT INTO fight_outcomes (fight_id, user_id, collectable_id, outcome) VALUES (?, ?, ?, ?);
`
)

func (sd *SDDB) CreateCombatReport(ctx context.Context, report *CombatReport) (*CombatReport, error) {
	tx, err := sd.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

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

	_, err = tx.ExecContext(ctx, insertCombatReportQuery, report.ID, participants, outcomes, knives, report.Event)
	if err != nil {
		return nil, err
	}

	for idx, user := range report.Participants {
		var knifeID *int64
		if idx < len(report.Knives) {
			knifeID = &report.Knives[idx]
		}

		_, err := tx.ExecContext(ctx, insertCombatOutcomeQuery, report.ID, user, knifeID, outcomeFromInt(report.Outcomes[idx]))
		if err != nil {
			return nil, fmt.Errorf("creating outcome entry: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("committing transaction: %w", err)
	}

	return sd.GetCombatReport(ctx, report.ID)
}

var getKnifeStatsForUserQuery = `
SELECT
  outcome,
  COUNT(*)
FROM fight_outcomes
WHERE user_id = ?
GROUP BY outcome;
`

func (sd *SDDB) GetCombatStatsForUser(ctx context.Context, userID int64) (CombatStats, error) {
	rows, err := sd.db.QueryContext(ctx, getKnifeStatsForUserQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("querying for knife stats: %w", err)
	}

	stats := make(map[string]int)

	for rows.Next() {
		var outcome string
		var count int

		err := rows.Scan(&outcome, &count)
		if err != nil {
			return nil, fmt.Errorf("scanning row: %w", err)
		}

		stats[outcome] = count
	}

	return stats, nil
}

var getKnifeStatsForKnifeQuery = `
SELECT
  outcome,
  COUNT(*)
FROM fight_outcomes
WHERE collectable_id = ?
GROUP BY outcome;
`

func (sd *SDDB) GetCombatStatsForKnife(ctx context.Context, knifeID int64) (CombatStats, error) {
	rows, err := sd.db.QueryContext(ctx, getKnifeStatsForKnifeQuery, knifeID)
	if err != nil {
		return nil, fmt.Errorf("querying for knife stats: %w", err)
	}

	stats := make(map[string]int)

	for rows.Next() {
		var outcome string
		var count int

		err := rows.Scan(&outcome, &count)
		if err != nil {
			return nil, fmt.Errorf("scanning row: %w", err)
		}

		stats[outcome] = count
	}

	return stats, nil
}

var getCombatReportsForEventQuery = `
SELECT
  id,
  participants,
  outcomes,
  knives,
  event,
  created_at
FROM fights
WHERE event = ?
ORDER BY created_at DESC;
`

func (sd *SDDB) GetCombatReportsForEvent(ctx context.Context, event string) ([]*CombatReport, error) {
	rows, err := sd.db.QueryContext(ctx, getCombatReportsForEventQuery, event)
	if err != nil {
		return nil, fmt.Errorf("querying for fights by event: %w", err)
	}

	reports := []*CombatReport{}

	for rows.Next() {
		var report CombatReport
		var participants string
		var outcomes string
		var knives string
		var createdAt string

		err := rows.Scan(
			&report.ID,
			&participants,
			&outcomes,
			&knives,
			&report.Event,
			&createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning row: %w", err)
		}

		err = json.Unmarshal([]byte(participants), &report.Participants)
		if err != nil {
			return nil, fmt.Errorf("decoding participants: %w", err)
		}
		err = json.Unmarshal([]byte(outcomes), &report.Outcomes)
		if err != nil {
			return nil, fmt.Errorf("decoding outcomes: %w", err)
		}
		err = json.Unmarshal([]byte(knives), &report.Knives)
		if err != nil {
			return nil, fmt.Errorf("decoding knives: %w", err)
		}
		report.CreatedAt, err = parseTimestamp(createdAt)
		if err != nil {
			return nil, fmt.Errorf("decoding createdAt: %w", err)
		}

		reports = append(reports, &report)
	}

	return reports, nil
}

func outcomeFromInt(outcome int) string {
	switch outcome {
	case 1:
		return "win"
	case 0:
		return "draw"
	case -1:
		return "loss"
	default:
		return ""
	}
}
