package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := sql.Open("mysql", os.Getenv("DSN"))
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}

	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping: %v", err)
	}

	q, err := db.Prepare("SELECT id FROM knives WHERE name = ?")
	if err != nil {
		log.Fatal(err)
	}

	wq, err := db.Prepare(`
UPDATE knives
SET image_name = ?
WHERE id = ?;
  `)
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Open("reference/knives.txt")
	if err != nil {
		log.Fatalf("unable to open file: %s", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Text()

		fields := strings.Split(line, "\t")

		if len(fields) < 3 {
			log.Fatalf("not enough fields on line: %s", line)
		}

		var id int
		r := q.QueryRow(fields[0])
		err := r.Scan(&id)
		if err != nil {
			log.Fatalf("looking up %s: %s", fields[0], err)
		}

		fmt.Printf("%d %s\n", id, path.Base(fields[2]))

		ext := path.Ext(fields[2])

		imageName := fmt.Sprintf("%d%s", id, ext)

		f, err := os.Create(fmt.Sprintf("reference/images/%s", imageName))
		if err != nil {
			log.Fatalf("creating file: %s", err)
		}

		src, err := os.Open(fmt.Sprintf("reference/knifezipo/%s", path.Base(fields[2])))
		if err != nil {
			log.Fatalf("opening src: %s", err)
		}
		_, err = io.Copy(f, src)
		if err != nil {
			log.Fatalf("copying: %s", err)
		}

		_, err = wq.Exec(imageName, id)
		if err != nil {
			log.Fatalf("updating image_name: %s", err)
		}
	}
}
