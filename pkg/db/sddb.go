package db

import (
	"context"
	"database/sql"
	"errors"
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
