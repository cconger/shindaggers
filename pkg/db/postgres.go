package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	model "github.com/cconger/shindaggers/pkg/db/.gen/postgres/public/model"
	table "github.com/cconger/shindaggers/pkg/db/.gen/postgres/public/table"
	postgres "github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
)

type Collection struct {
	ID      int64
	Creator *User
	Name    string

	CreatedAt time.Time
	EnabledAt time.Time
	RetiredAt time.Time
}

type Collectable struct {
	model.Collectables `alias:"collectable"`

	Creator    *model.Users `alias:"creator"`
	ApprovedBy *model.Users `alias:"approver"`
}

type CollectableInstance struct {
	model.CollectableInstances

	Collectable *Collectable `alias:"collectable"`
	Owner       *model.Users `alias:"owner"`
	Edition     *model.Editions
}

type PostgresDB struct {
	DB *sql.DB
}

type GetLatestIssuesOptions struct {
	ByCreator    int64
	ByCollection int64
	After        time.Time
}

type User struct {
	model.Users
}

type ConstraintBuilder struct {
	constraints []postgres.BoolExpression
}

func (c *ConstraintBuilder) Add(b postgres.BoolExpression) {
	c.constraints = append(c.constraints, b)
}

func (c *ConstraintBuilder) Apply(stmt postgres.SelectStatement) postgres.SelectStatement {
	switch len(c.constraints) {
	case 0:
		return stmt
	case 1:
		return stmt.WHERE(c.constraints[0])
	default:
		return stmt.WHERE(postgres.AND(c.constraints...))
	}
}

func (db *PostgresDB) GetLatestIssues(ctx context.Context, options GetLatestIssuesOptions) ([]CollectableInstance, error) {
	owner := table.Users.AS("owner")
	creator := table.Users.AS("creator")
	collectable := table.Collectables.AS("collectable")

	where := []postgres.BoolExpression{
		collectable.DeletedAt.IS_NULL(),
	}
	if !options.After.IsZero() {
		where = append(where, table.CollectableInstances.CreatedAt.GT(postgres.TimestampT(options.After)))
	}
	if options.ByCollection != 0 {
		where = append(where, collectable.CollectionID.EQ(postgres.Int64(options.ByCollection)))
	}

	if len(where) == 0 {
		where = append(where, postgres.Bool(true))
	}

	stmt := postgres.SELECT(
		table.CollectableInstances.AllColumns,
		collectable.AllColumns,
		owner.AllColumns.Except(owner.Admin, owner.CreatedAt),
		creator.AllColumns.Except(creator.Admin, creator.CreatedAt),
		table.Editions.AllColumns,
	).FROM(
		table.CollectableInstances.
			INNER_JOIN(collectable, table.CollectableInstances.CollectableID.EQ(collectable.ID)).
			INNER_JOIN(creator, collectable.CreatorID.EQ(creator.ID)).
			INNER_JOIN(owner, table.CollectableInstances.OwnerID.EQ(owner.ID)).
			INNER_JOIN(table.Editions, table.CollectableInstances.EditionID.EQ(table.Editions.ID)),
	).WHERE(
		postgres.AND(where...),
	).ORDER_BY(
		table.CollectableInstances.CreatedAt.DESC(),
	).LIMIT(15)

	dest := []CollectableInstance{}

	err := stmt.QueryContext(ctx, db.DB, &dest)
	if err != nil {
		return nil, err
	}

	return dest, nil
}

func (db *PostgresDB) CreateCollectableInstance(ctx context.Context, instance model.CollectableInstances) (*CollectableInstance, error) {
	stmt := table.CollectableInstances.
		INSERT(table.CollectableInstances.AllColumns).
		MODEL(instance).
		RETURNING(table.CollectableInstances.ID)

	id := []int64{}
	err := stmt.QueryContext(ctx, db.DB, &id)
	if err != nil {
		return nil, err
	}

	if len(id) != 1 {
		return nil, errors.New("expected one id to be returned")
	}

	instances, err := db.GetCollectableInstances(ctx, GetCollectableInstancesOptions{ByID: id[0]})
	if err != nil {
		return nil, err
	}
	if len(instances) == 0 {
		return nil, ErrNotFound
	}

	return &instances[0], nil
}

type GetCollectableInstancesOptions struct {
	ByOwner       int64
	ByID          int64
	ByCollectable int64
	GetDeleted    bool
}

func (db *PostgresDB) GetCollectableInstances(ctx context.Context, options GetCollectableInstancesOptions) ([]CollectableInstance, error) {
	collectable := table.Collectables.AS("collectable")
	creator := table.Users.AS("creator")
	owner := table.Users.AS("owner")

	stmt := postgres.SELECT(
		table.CollectableInstances.AllColumns,
		collectable.AllColumns,
		creator.AllColumns.Except(creator.Admin, creator.CreatedAt),
		owner.AllColumns.Except(owner.Admin, owner.CreatedAt),
		table.Editions.AllColumns,
	).FROM(
		table.CollectableInstances.
			INNER_JOIN(collectable, table.CollectableInstances.CollectableID.EQ(collectable.ID)).
			INNER_JOIN(creator, collectable.CreatorID.EQ(creator.ID)).
			INNER_JOIN(owner, table.CollectableInstances.OwnerID.EQ(owner.ID)).
			INNER_JOIN(table.Editions, table.CollectableInstances.EditionID.EQ(table.Editions.ID)),
	)

	c := ConstraintBuilder{}

	if options.ByOwner != 0 {
		c.Add(owner.ID.EQ(postgres.Int64(options.ByOwner)))
	}
	if options.ByCollectable != 0 {
		c.Add(collectable.ID.EQ(postgres.Int64(options.ByCollectable)))
	}
	if options.ByID != 0 {
		c.Add(table.CollectableInstances.ID.EQ(postgres.Int64(options.ByID)))
	}
	if !options.GetDeleted {
		c.Add(table.CollectableInstances.DeletedAt.IS_NULL())
	}

	stmt = c.Apply(stmt)

	stmt.ORDER_BY(table.CollectableInstances.CreatedAt.DESC())

	dest := []CollectableInstance{}
	err := stmt.QueryContext(ctx, db.DB, &dest)
	if err != nil {
		return nil, err
	}

	return dest, nil
}

func (db *PostgresDB) GetEquippedForUser(ctx context.Context, userID int64) (*CollectableInstance, error) {
	collectable := table.Collectables.AS("collectable")
	creator := table.Users.AS("creator")
	owner := table.Users.AS("owner")

	stmt := postgres.SELECT(
		table.CollectableInstances.AllColumns,
		collectable.AllColumns,
		creator.AllColumns.Except(creator.Admin, creator.CreatedAt),
		owner.AllColumns.Except(owner.Admin, owner.CreatedAt),
		table.Editions.AllColumns,
	).FROM(
		table.UserEquipCollectableInstance.
			INNER_JOIN(table.CollectableInstances, table.UserEquipCollectableInstance.InstanceID.EQ(table.CollectableInstances.ID)).
			INNER_JOIN(collectable, table.CollectableInstances.CollectableID.EQ(collectable.ID)).
			INNER_JOIN(creator, collectable.CreatorID.EQ(creator.ID)).
			INNER_JOIN(owner, table.CollectableInstances.OwnerID.EQ(owner.ID)).
			INNER_JOIN(table.Editions, table.CollectableInstances.EditionID.EQ(table.Editions.ID)),
	).WHERE(
		table.UserEquipCollectableInstance.UserID.EQ(postgres.Int64(userID)),
	).LIMIT(1)

	dest := CollectableInstance{}
	err := stmt.QueryContext(ctx, db.DB, &dest)
	if err != nil {
		return nil, err
	}

	return &dest, nil
}

type GetCollectableOptions struct {
	Name          string
	GetDeleted    bool
	GetUnapproved bool
}

func (db *PostgresDB) GetCollectable(ctx context.Context, id int64, options GetCollectableOptions) (*Collectable, error) {
	collectable := table.Collectables.AS("collectable")
	creator := table.Users.AS("creator")

	stmt := postgres.SELECT(
		collectable.AllColumns,
		creator.AllColumns,
	).FROM(
		collectable.INNER_JOIN(creator, collectable.CreatorID.EQ(creator.ID)),
	)

	c := ConstraintBuilder{}
	c.Add(collectable.ID.EQ(postgres.Int64(id)))

	if !options.GetDeleted {
		c.Add(collectable.DeletedAt.IS_NULL())
	}
	if !options.GetUnapproved {
		c.Add(collectable.ApprovedAt.IS_NOT_NULL())
	}
	if options.Name != "" {
		c.Add(collectable.Name.EQ(postgres.String(options.Name)))
	}

	stmt = c.Apply(stmt).LIMIT(1)

	dest := Collectable{}

	err := stmt.QueryContext(ctx, db.DB, &dest)
	if err != nil {
		return nil, err
	}

	return &dest, nil
}

type GetCollectablesOptions struct {
	Collection     int64
	Creator        int64
	Rarity         string
	GetDeleted     bool
	GetUnapproved  bool
	OnlyUnapproved bool
}

func (db *PostgresDB) GetCollectables(ctx context.Context, options GetCollectablesOptions) ([]*Collectable, error) {
	collectable := table.Collectables.AS("collectable")
	creator := table.Users.AS("creator")

	stmt := postgres.SELECT(
		collectable.AllColumns,
		creator.AllColumns,
	).FROM(
		collectable.INNER_JOIN(creator, collectable.CreatorID.EQ(creator.ID)),
	)

	c := ConstraintBuilder{}
	if options.Collection != 0 {
		c.Add(collectable.CollectionID.EQ(postgres.Int64(options.Collection)))
	}
	if options.Creator != 0 {
		c.Add(creator.ID.EQ(postgres.Int64(options.Creator)))
	}
	if options.Rarity != "" {
		c.Add(collectable.Rarity.EQ(postgres.String(options.Rarity)))
	}
	if !options.GetDeleted {
		c.Add(collectable.DeletedAt.IS_NULL())
	}
	if !options.GetUnapproved {
		c.Add(collectable.ApprovedAt.IS_NOT_NULL())
	}
	if options.OnlyUnapproved {
		c.Add(collectable.ApprovedAt.IS_NULL())
	}
	stmt = c.Apply(stmt)

	stmt.ORDER_BY(collectable.CreatedAt.ASC())

	dest := []*Collectable{}

	err := stmt.QueryContext(ctx, db.DB, &dest)
	if err != nil {
		return nil, err
	}

	return dest, nil
}

func (db *PostgresDB) SearchUsers(ctx context.Context, search string) ([]User, error) {
	stmt := table.Users.SELECT(
		table.Users.AllColumns,
	).
		FROM(table.Users).
		WHERE(table.Users.Name.REGEXP_LIKE(postgres.String(search), false))

	dest := []User{}
	err := stmt.QueryContext(ctx, db.DB, &dest)
	if err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return dest, nil
}

type GetUserOptions struct {
	ID        int64
	TwitchID  string
	Username  string
	AuthToken []byte
}

func (db *PostgresDB) GetUser(ctx context.Context, options GetUserOptions) (*User, error) {
	stmt := postgres.SELECT(
		table.Users.AllColumns,
	)

	if options.AuthToken != nil {
		stmt = stmt.FROM(
			table.UserTokens.INNER_JOIN(table.Users, table.UserTokens.UserID.EQ(table.Users.ID)),
		).WHERE(
			table.UserTokens.Token.EQ(postgres.Bytea(options.AuthToken)),
		).LIMIT(1)
	} else {
		stmt = stmt.FROM(table.Users)

		c := ConstraintBuilder{}

		if options.ID != 0 {
			c.Add(table.Users.ID.EQ(postgres.Int64(options.ID)))
		}
		if options.TwitchID != "" {
			c.Add(table.Users.TwitchID.EQ(postgres.String(options.TwitchID)))
		}
		if options.Username != "" {
			c.Add(table.Users.Name.EQ(postgres.String(options.Username)))
		}
		stmt = c.Apply(stmt).LIMIT(1)
	}

	dest := User{}
	err := stmt.QueryContext(ctx, db.DB, &dest)
	if err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &dest, nil
}

func (db *PostgresDB) CreateUser(ctx context.Context, user User) (*User, error) {
	stmt := table.Users.INSERT(
		table.Users.AllColumns,
	).MODEL(user.Users).RETURNING(table.Users.AllColumns)

	dest := User{}
	err := stmt.QueryContext(ctx, db.DB, &dest)
	if err != nil {
		return nil, err
	}

	return &dest, nil
}

func (db *PostgresDB) UpdateUser(ctx context.Context, user User) (*User, error) {
	stmt := table.Users.UPDATE(
		table.Users.AllColumns.Except(table.Users.ID),
	).MODEL(user.Users).
		WHERE(table.Users.ID.EQ(postgres.Int64(user.ID))).
		RETURNING(table.Users.AllColumns)

	dest := User{}
	err := stmt.QueryContext(ctx, db.DB, &dest)
	if err != nil {
		return nil, err
	}

	return &dest, nil
}

func (db *PostgresDB) SetEquipped(ctx context.Context, collectableInstanceID int64, userID int64) error {
	now := time.Now()
	stmt := table.UserEquipCollectableInstance.
		INSERT(
			table.UserEquipCollectableInstance.AllColumns,
		).MODEL(model.UserEquipCollectableInstance{
		UserID:     userID,
		InstanceID: &collectableInstanceID,
		EquippedAt: &now,
	}).
		ON_CONFLICT(table.UserEquipCollectableInstance.UserID).
		DO_UPDATE(
			postgres.SET(
				table.UserEquipCollectableInstance.InstanceID.SET(table.UserEquipCollectableInstance.EXCLUDED.InstanceID),
				table.UserEquipCollectableInstance.EquippedAt.SET(table.UserEquipCollectableInstance.EXCLUDED.EquippedAt),
			),
		)

	_, err := stmt.ExecContext(ctx, db.DB)
	if err != nil {
		return err
	}
	return nil
}

func (db *PostgresDB) CreateCollectable(ctx context.Context, collectable model.Collectables) (*Collectable, error) {
	stmt := table.Collectables.INSERT(
		table.Collectables.AllColumns.Except(table.Collectables.CreatedAt),
	).
		MODEL(collectable).
		RETURNING(table.Collectables.AllColumns)

	dest := model.Collectables{}
	err := stmt.QueryContext(ctx, db.DB, &dest)
	if err != nil {
		return nil, err
	}

	return db.GetCollectable(ctx, dest.ID, GetCollectableOptions{
		GetUnapproved: true,
	})
}

func (db *PostgresDB) UpdateCollectable(ctx context.Context, collectable model.Collectables) (*Collectable, error) {
	stmt := table.Collectables.
		UPDATE(
			table.Collectables.Name,
			table.Collectables.CreatorID,
			table.Collectables.Rarity,
			table.Collectables.Imagepath,
		).
		MODEL(collectable).
		WHERE(table.Collectables.ID.EQ(postgres.Int64(collectable.ID))).
		RETURNING(table.Collectables.AllColumns)

	dest := model.Collectables{}
	err := stmt.QueryContext(ctx, db.DB, &dest)
	if err != nil {
		return nil, err
	}

	return db.GetCollectable(ctx, dest.ID, GetCollectableOptions{
		GetUnapproved: true,
	})
}

func (db *PostgresDB) ApproveCollectable(ctx context.Context, collectableID int64, approverID int64) (*Collectable, error) {
	stmt := table.Collectables.
		UPDATE(
			table.Collectables.ApprovedAt,
			table.Collectables.ApprovedBy,
		).
		SET(
			table.Collectables.ApprovedAt.SET(postgres.TimestampT(time.Now())),
			table.Collectables.ApprovedBy.SET(postgres.Int64(approverID)),
		).
		WHERE(table.Collectables.ID.EQ(postgres.Int64(collectableID))).
		RETURNING(table.Collectables.AllColumns)

	dest := model.Collectables{}
	err := stmt.QueryContext(ctx, db.DB, &dest)
	if err != nil {
		return nil, err
	}

	return db.GetCollectable(ctx, dest.ID, GetCollectableOptions{
		GetUnapproved: true,
	})
}

func (db *PostgresDB) DeleteCollectable(ctx context.Context, collectableID int64) error {
	stmt := table.Collectables.
		UPDATE(
			table.Collectables.DeletedAt,
		).
		SET(
			table.Collectables.DeletedAt.SET(postgres.TimestampT(time.Now())),
		).
		WHERE(table.Collectables.ID.EQ(postgres.Int64(collectableID))).
		RETURNING(table.Collectables.AllColumns)

	_, err := stmt.ExecContext(ctx, db.DB)
	if err != nil {
		return err
	}

	return nil
}

func (db *PostgresDB) CreateImageUpload(ctx context.Context, imageID int64, user int64, name, uploadname string) error {
	stmt := table.ImageUploads.INSERT(
		table.ImageUploads.AllColumns,
	).MODEL(model.ImageUploads{
		ID:         imageID,
		UserID:     user,
		Imagepath:  name,
		UploadName: &uploadname,
	})

	_, err := stmt.ExecContext(ctx, db.DB)
	if err != nil {
		return err
	}
	return nil
}

func (db *PostgresDB) SaveAuth(ctx context.Context, auth UserAuth) error {
	m := model.UserTokens{
		UserID:       auth.UserID,
		Token:        auth.Token,
		AccessToken:  &auth.AccessToken,
		RefreshToken: &auth.RefreshToken,
		ExpiresAt:    &auth.ExpiresAt,
	}

	stmt := table.UserTokens.INSERT(table.UserTokens.AllColumns).MODEL(m).ON_CONFLICT(table.UserTokens.Token).DO_UPDATE(postgres.SET(
		table.UserTokens.Token.SET(table.UserTokens.EXCLUDED.Token),
		table.UserTokens.UserID.SET(table.UserTokens.EXCLUDED.UserID),
		table.UserTokens.AccessToken.SET(table.UserTokens.EXCLUDED.AccessToken),
		table.UserTokens.RefreshToken.SET(table.UserTokens.EXCLUDED.RefreshToken),
		table.UserTokens.ExpiresAt.SET(table.UserTokens.EXCLUDED.ExpiresAt),
	))

	_, err := stmt.ExecContext(ctx, db.DB)
	if err != nil {
		return err
	}

	return nil
}

func (db *PostgresDB) GetWeights(ctx context.Context, collectionID int64) ([]*PullWeight, error) {
	stmt := table.Collections.
		SELECT(table.Collections.Weights).
		FROM(table.Collections).
		WHERE(table.Collections.ID.EQ(postgres.Int64(collectionID))).
		LIMIT(1)

	dest := []string{}
	err := stmt.QueryContext(ctx, db.DB, &dest)
	if err != nil {
		return nil, err
	}

	if len(dest) != 1 {
		return nil, ErrNotFound
	}

	m := make(map[string]int)
	err = json.Unmarshal([]byte(dest[0]), &m)
	if err != nil {
		return nil, err
	}

	out := []*PullWeight{}
	for k, v := range m {
		out = append(out, &PullWeight{
			Rarity: k,
			Weight: v,
		})
	}

	return out, nil
}
