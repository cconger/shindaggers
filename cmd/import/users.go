package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/cconger/shindaggers/pkg/twitch"

	_ "github.com/go-sql-driver/mysql"
)

var usersQuery = `
SELECT 
  twitch_name,
  lookup_name
FROM users
WHERE twitch_id is NULL
LIMIT 100;
`

var updateUser = `
UPDATE users
SET twitch_name = ?, twitch_id = ?
WHERE lookup_name = ?;
`

func importUsers() error {
	// Query for all users that don't have a twitchid and lets add them...

	log.Println("importUsers")

	twitchClientID := os.Getenv("TWITCH_CLIENT_ID")
	twitchSecret := os.Getenv("TWITCH_SECRET")

	client, err := twitch.NewClient(twitchClientID, twitchSecret)
	if err != nil {
		return fmt.Errorf("could not create twitch client: %w", err)
	}

	ctx := context.Background()
	token, err := client.GetAppToken(ctx)
	if err != nil {
		return fmt.Errorf("getting app token: %w", err)
	}

	db, err := sql.Open("mysql", os.Getenv("DSN"))
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		return err
	}

	// run usersQuery over and over until you get no results
	q, err := db.Prepare(usersQuery)
	if err != nil {
		return fmt.Errorf("prepping null users query: %w", err)
	}

	rows, err := q.Query()
	if err != nil {
		return fmt.Errorf("querying null users: %w", err)
	}

	for rows.Next() {
		var name string
		var login string
		rows.Scan(
			&name,
			&login,
		)

		if login == "test_user" {
			continue
		}

		users, err := client.GetUsersByLogin(ctx, token.AccessToken, login)
		if err != nil {
			log.Printf("error getting %s: %s", login, err)
			continue
		}

		if len(users) < 1 {
			log.Printf("empty result for %s", login)
			continue
		}

		u := users[0]

		log.Printf("Updating %s %s %s", u.DisplayName, u.ID, u.Login)
		uq, err := db.Prepare(updateUser)
		if err != nil {
			return fmt.Errorf("preparing update: %w", err)
		}
		_, err = uq.Exec(u.DisplayName, u.ID, u.Login)
		if err != nil {
			return fmt.Errorf("updating: %w", err)
		}
	}

	return nil
}
