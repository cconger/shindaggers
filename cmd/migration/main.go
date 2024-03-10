package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	pgmodel "github.com/cconger/shindaggers/pkg/db/.gen/postgres/public/model"
	pgtable "github.com/cconger/shindaggers/pkg/db/.gen/postgres/public/table"
	sdmodel "github.com/cconger/shindaggers/pkg/db/.gen/sd/model"
	sdtable "github.com/cconger/shindaggers/pkg/db/.gen/sd/table"
	mysql "github.com/go-jet/jet/v2/mysql"

	// postgres "github.com/go-jet/jet/v2/postgres"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// This is a tool to migrate from planetscale to supabase
func main() {
	sddb, err := sql.Open("mysql", os.Getenv("DSN"))
	if err != nil {
		panic(err)
	}

	pgdb, err := sql.Open("pgx", os.Getenv("SUPABASE_DSN"))
	if err != nil {
		panic(err)
	}

	err = migrateUsers(sddb, pgdb)
	if err != nil {
		panic(err)
	}

	err = migrateUserTokens(sddb, pgdb)
	if err != nil {
		panic(err)
	}

	err = migrateImageUploads(sddb, pgdb)
	if err != nil {
		panic(err)
	}

	err = createDefaultCollection(pgdb)
	if err != nil {
		panic(err)
	}

	err = migrateCollectables(sddb, pgdb)
	if err != nil {
		panic(err)
	}

	err = migrateEditions(sddb, pgdb)
	if err != nil {
		panic(err)
	}

	err = migrateCollectableInstances(sddb, pgdb)
	if err != nil {
		panic(err)
	}

	err = migrateEquipped(sddb, pgdb)
	if err != nil {
		panic(err)
	}
}

func migrateUsers(sddb *sql.DB, pgdb *sql.DB) error {
	q := mysql.SELECT(
		sdtable.Users.AllColumns,
	).FROM(
		sdtable.Users,
	)

	res := []sdmodel.Users{}

	err := q.Query(sddb, &res)
	if err != nil {
		return fmt.Errorf("failed to query users: %w", err)
	}

	pgmodels := make([]pgmodel.Users, len(res))
	for idx, user := range res {
		pgmodels[idx] = pgmodel.Users{
			ID:        user.ID,
			TwitchID:  user.TwitchID,
			Name:      *user.TwitchName,
			CreatedAt: *user.CreatedAt,
		}
	}

	insert := pgtable.Users.INSERT(pgtable.Users.AllColumns).MODELS(pgmodels).ON_CONFLICT(pgtable.Users.ID).DO_NOTHING()
	r, err := insert.Exec(pgdb)
	if err != nil {
		return fmt.Errorf("failed to insert users: %w", err)
	}

	rows, err := r.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	fmt.Printf("inserted %d users\n", rows)

	return nil
}

func migrateUserTokens(sddb *sql.DB, pgdb *sql.DB) error {
	q := mysql.SELECT(
		sdtable.UserAuth.AllColumns,
	).FROM(
		sdtable.UserAuth,
	)

	res := []sdmodel.UserAuth{}

	err := q.Query(sddb, &res)
	if err != nil {
		return fmt.Errorf("failed to query user_tokens: %w", err)
	}

	pgmodels := make([]pgmodel.UserTokens, len(res))
	for idx, ua := range res {
		pgmodels[idx] = pgmodel.UserTokens{
			UserID:       *ua.UserID,
			Token:        ua.Token,
			AccessToken:  ua.AccessToken,
			RefreshToken: ua.RefreshToken,
			ExpiresAt:    ua.ExpiresAt,
			CreatedAt:    *ua.CreatedAt,
			UpdatedAt:    *ua.UpdatedAt,
		}
	}

	insert := pgtable.UserTokens.INSERT(
		pgtable.UserTokens.AllColumns,
	).MODELS(
		pgmodels,
	).ON_CONFLICT(pgtable.UserTokens.Token).DO_NOTHING()
	r, err := insert.Exec(pgdb)
	if err != nil {
		return fmt.Errorf("failed to insert user_tokens: %w", err)
	}

	rows, err := r.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	fmt.Printf("inserted %d user_tokens\n", rows)

	return nil
}

func migrateImageUploads(sddb *sql.DB, pgdb *sql.DB) error {
	q := mysql.SELECT(
		sdtable.ImageUploads.AllColumns,
	).FROM(
		sdtable.ImageUploads,
	)

	res := []sdmodel.ImageUploads{}

	err := q.Query(sddb, &res)
	if err != nil {
		return fmt.Errorf("failed to query image_uploads: %w", err)
	}

	pgmodels := make([]pgmodel.ImageUploads, len(res))
	for idx, ia := range res {
		pgmodels[idx] = pgmodel.ImageUploads{
			ID:         ia.ImageID,
			UploadName: ia.Uploadname,
			UserID:     *ia.UserID,
			Imagepath:  *ia.Path,
			UploadedAt: *ia.UploadedAt,
		}
	}

	insert := pgtable.ImageUploads.INSERT(
		pgtable.ImageUploads.AllColumns,
	).MODELS(
		pgmodels,
	).ON_CONFLICT(pgtable.ImageUploads.ID).DO_NOTHING()
	r, err := insert.Exec(pgdb)
	if err != nil {
		return fmt.Errorf("failed to insert image_uploads: %w", err)
	}

	rows, err := r.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	fmt.Printf("inserted %d image_uploads\n", rows)

	return nil
}

var weights string = `
{
  "Uncommon": 200,
  "Common": 400,
  "Rare": 250,
  "Ultra Rare": 50,
  "Super Rare": 100
}
`

func createDefaultCollection(pgdb *sql.DB) error {
	now := time.Now().UTC()
	collection := pgmodel.Collections{
		ID:        1,
		Name:      "Shindaggers",
		Weights:   &weights,
		CreatorID: 6,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		ActiveAt:  &now,
	}

	insert := pgtable.Collections.INSERT(pgtable.Collections.AllColumns).MODEL(collection).ON_CONFLICT(pgtable.Collections.ID).DO_NOTHING()

	r, err := insert.Exec(pgdb)
	if err != nil {
		return fmt.Errorf("failed to insert default collection: %w", err)
	}

	rows, err := r.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	fmt.Printf("inserted %d collections\n", rows)

	return nil
}

func migrateEditions(sddb *sql.DB, pgdb *sql.DB) error {
	q := mysql.SELECT(
		sdtable.Editions.AllColumns,
	).FROM(
		sdtable.Editions,
	)

	res := []sdmodel.Editions{}

	err := q.Query(sddb, &res)
	if err != nil {
		return fmt.Errorf("failed to query editions: %w", err)
	}

	pgmodels := make([]pgmodel.Editions, len(res))
	for idx, ed := range res {
		pgmodels[idx] = pgmodel.Editions{
			ID:        int64(ed.ID),
			Name:      *ed.Name,
			CreatedAt: *ed.CreatedAt,
		}
	}

	insert := pgtable.Editions.INSERT(
		pgtable.Editions.AllColumns,
	).MODELS(
		pgmodels,
	).ON_CONFLICT(pgtable.Editions.ID).DO_NOTHING()
	r, err := insert.Exec(pgdb)
	if err != nil {
		return fmt.Errorf("failed to insert editions: %w", err)
	}

	rows, err := r.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	fmt.Printf("inserted %d editions\n", rows)

	return nil
}

func migrateCollectables(sddb *sql.DB, pgdb *sql.DB) error {
	q := mysql.SELECT(
		sdtable.Knives.AllColumns,
	).FROM(
		sdtable.Knives,
	)

	res := []sdmodel.Knives{}

	err := q.Query(sddb, &res)
	if err != nil {
		return fmt.Errorf("failed to query knives: %w", err)
	}

	collection := int64(1)

	pgmodels := make([]pgmodel.Collectables, len(res))
	for idx, c := range res {
		var dat *time.Time = nil
		if c.Deleted != nil && *c.Deleted {
			now := time.Now().UTC()
			dat = &now
		}
		pgmodels[idx] = pgmodel.Collectables{
			ID:           c.ID,
			Name:         *c.Name,
			CollectionID: &collection,
			CreatorID:    *c.AuthorID,
			Rarity:       c.Rarity.String(),
			Imagepath:    *c.ImageName,
			CreatedAt:    *c.CreatedAt,
			ApprovedAt:   c.ApprovedAt,
			ApprovedBy:   c.ApprovedBy,
			DeletedAt:    dat,
		}
	}

	insert := pgtable.Collectables.INSERT(
		pgtable.Collectables.AllColumns,
	).MODELS(
		pgmodels,
	).ON_CONFLICT(pgtable.Collectables.ID).DO_NOTHING()
	r, err := insert.Exec(pgdb)
	if err != nil {
		return fmt.Errorf("failed to insert collectables: %w", err)
	}

	rows, err := r.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	fmt.Printf("inserted %d collectables\n", rows)

	return nil
}

type Tags struct {
	Subscriber bool `json:"subscriber"`
	Verified   bool `json:"verified"`
}

func migrateCollectableInstances(sddb *sql.DB, pgdb *sql.DB) error {
	offset := int64(0)
	for {
		q := mysql.SELECT(
			sdtable.KnifeOwnership.AllColumns,
		).FROM(
			sdtable.KnifeOwnership,
		).ORDER_BY(sdtable.KnifeOwnership.InstanceID).LIMIT(500).OFFSET(offset)

		res := []sdmodel.KnifeOwnership{}

		err := q.Query(sddb, &res)
		if err != nil {
			return fmt.Errorf("failed to query knive_ownership: %w", err)
		}

		count := len(res)
		if count == 0 {
			fmt.Println("no more results")
			return nil
		}
		fmt.Println("got", count, "results")
		offset += int64(count)

		pgmodels := make([]pgmodel.CollectableInstances, len(res))
		for idx, k := range res {
			t := Tags{}
			if k.IsVerified != nil && *k.IsVerified {
				t.Verified = true
			}

			if k.WasSubscriber != nil && *k.WasSubscriber {
				t.Subscriber = true
			}

			tv, err := json.Marshal(&t)
			if err != nil {
				return fmt.Errorf("unable to encode tags: %w", err)
			}
			tagString := string(tv)

			pgmodels[idx] = pgmodel.CollectableInstances{
				ID:            k.InstanceID,
				CollectableID: *k.KnifeID,
				OwnerID:       *k.UserID,
				EditionID:     1,
				CreatedAt:     *k.TransactedAt,
				Tags:          &tagString,
			}
		}

		insert := pgtable.CollectableInstances.INSERT(
			pgtable.CollectableInstances.AllColumns,
		).MODELS(
			pgmodels,
		).ON_CONFLICT(pgtable.CollectableInstances.ID).DO_NOTHING()
		r, err := insert.Exec(pgdb)
		if err != nil {
			return fmt.Errorf("failed to insert collectable_instances: %w", err)
		}

		rows, err := r.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}
		fmt.Printf("inserted %d collectable_instances\n", rows)
	}
}

func migrateEquipped(sddb *sql.DB, pgdb *sql.DB) error {
	q := mysql.SELECT(
		sdtable.Equipped.AllColumns,
	).FROM(
		sdtable.Equipped,
	)

	res := []sdmodel.Equipped{}

	err := q.Query(sddb, &res)
	if err != nil {
		return fmt.Errorf("failed to query equipped: %w", err)
	}

	pgmodels := make([]pgmodel.UserEquipCollectableInstance, len(res))
	for idx, c := range res {
		pgmodels[idx] = pgmodel.UserEquipCollectableInstance{
			UserID:     c.UserID,
			InstanceID: c.InstanceID,
			EquippedAt: c.EquippedAt,
		}
	}

	insert := pgtable.UserEquipCollectableInstance.INSERT(
		pgtable.UserEquipCollectableInstance.AllColumns,
	).MODELS(
		pgmodels,
	).ON_CONFLICT(pgtable.UserEquipCollectableInstance.UserID).DO_NOTHING()
	r, err := insert.Exec(pgdb)
	if err != nil {
		return fmt.Errorf("failed to insert equipped: %w", err)
	}

	rows, err := r.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	fmt.Printf("inserted %d equipped\n", rows)

	return nil
}
