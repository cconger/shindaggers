package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const timeLayout = "2006-01-02 15:04:05"

type pull struct {
	time          time.Time
	username      string
	knife         string
	creator       string
	rarity        string
	verified      bool
	subscriber    bool
	first_edition bool
}

func mustBool(s string) bool {
	if s == "" {
		return false
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		log.Fatalf("failed to parse int from string (%s) %s", s, err)
	}
	return i > 0
}

func main() {
	// Open bladechain.txt and parse each line into a pull

	f, err := os.Open("bladechain.txt")
	if err != nil {
		log.Fatalf("Unable to open file: %s", err)
	}
	defer f.Close()

	// Create a slice of pulls
	pulls := make([]pull, 0)

	// Create a scanner to parse the file line by line
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		// Get the line from the scanner
		line := scanner.Text()

		// Split the line into a slice of strings
		fields := strings.Split(line, "\t")

		if len(fields) <= 7 {
			log.Fatalf("unable to parse line, not enough values: %s", line)
		}

		t, err := time.Parse(timeLayout, fields[0])
		if err != nil {
			log.Fatalf("unable to parse timestamp: %s - %s", fields[0], err)
		}
		// Create a pull from the fields
		p := pull{
			time:          t,
			username:      fields[1],
			knife:         fields[2],
			creator:       fields[3],
			rarity:        fields[4],
			verified:      mustBool(fields[5]),
			subscriber:    mustBool(fields[6]),
			first_edition: mustBool(fields[7]),
		}

		// Append the pull to the slice of pulls
		pulls = append(pulls, p)
	}

	// Check for errors
	if err := scanner.Err(); err != nil {
		log.Fatalf("Error scanning file: %s", err)
	}

	log.Printf("Parsed %d pulls", len(pulls))

	db, err := sql.Open("mysql", os.Getenv("DSN"))
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}

	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping: %v", err)
	}

	knivesByName := make(map[string]int)
	usersByName := make(map[string]int)

	insertPullQuery, err := db.Prepare("INSERT INTO knife_ownership (user_id, knife_id, trans_type) VALUES (?, ?,?);")
	if err != nil {
		log.Fatalf("unable to prepare knife creation query: %s", err)
	}

	// TODO: query for the max timestamp and only ingest ones after that time

	for _, p := range pulls {
		creator, ok := usersByName[p.creator]
		if !ok {
			id, err := getOrCreateUserIDByName(db, p.creator, p.time)
			if err != nil {
				log.Fatalf("could not create creator: %s", err)
			}
			usersByName[p.creator] = id
			creator = id
		}

		user, ok := usersByName[p.username]
		if !ok {
			id, err := getOrCreateUserIDByName(db, p.username, p.time)
			if err != nil {
				log.Fatalf("could not resolve knife by name: %s", err)
			}
			usersByName[p.username] = id
			user = id
		}

		knife, ok := knivesByName[p.knife]
		if !ok {
			id, err := getOrCreateKnifeIDByName(db, p.knife, p.rarity, creator, p.time)
			if err != nil {
				log.Fatalf("could not resolve knife by name: %s", err)
			}
			knivesByName[p.knife] = id
			knife = id
		}

		res, err := insertPullQuery.Exec(user, knife, "pull")
		if err != nil {
			log.Fatalf("unable to create pull: %s", err)
		}
		id, err := res.LastInsertId()
		if err != nil {
			log.Fatalf("unable to get id for created pull: %s", err)
		}
		log.Printf("%+v saved to %d\n\n", p, id)
	}
}

func getOrCreateKnifeIDByName(db *sql.DB, name string, rarity string, author_id int, trans_time time.Time) (int, error) {
	getKnifeQuery, err := db.Prepare("SELECT id FROM knives WHERE name = ?")
	if err != nil {
		log.Fatal(err)
	}

	rows, err := getKnifeQuery.Query(name)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	if !rows.Next() {
		// Insert knife
		createKnifeQuery, err := db.Prepare("INSERT INTO knives (name, author_id, rarity, edition_id, created_at) VALUES(?, ?, ?, ?, ?)")
		if err != nil {
			return 0, fmt.Errorf("unable to prepare insert query: %w", err)
		}

		res, err := createKnifeQuery.Exec(name, author_id, rarity, 1, trans_time)
		if err != nil {
			return 0, err
		}

		id, err := res.LastInsertId()
		return int(id), err
	}

	var id int
	err = rows.Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func getOrCreateUserIDByName(db *sql.DB, name string, created_at time.Time) (int, error) {
	getUserQuery, err := db.Prepare("SELECT id FROM users WHERE twitch_name = ?")
	if err != nil {
		log.Fatal(err)
	}

	rows, err := getUserQuery.Query(name)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	if !rows.Next() {
		// Insert knife
		createUserQuery, err := db.Prepare("INSERT INTO users (twitch_name, created_at) VALUES(?, ?)")
		if err != nil {
			return 0, fmt.Errorf("unable to prepare insert query: %w", err)
		}

		res, err := createUserQuery.Exec(name, created_at)
		if err != nil {
			return 0, err
		}
		id, err := res.LastInsertId()
		return int(id), err
	}

	var id int
	err = rows.Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}
