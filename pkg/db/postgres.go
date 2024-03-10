package db

import (
	"context"
	"database/sql"
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
		MODEL(instance)

	r, err := stmt.ExecContext(ctx, db.DB)
	if err != nil {
		return nil, err
	}

	id, err := r.LastInsertId()
	if err != nil {
		return nil, err
	}

	instances, err := db.GetCollectableInstances(ctx, GetCollectableInstancesOptions{ByID: id})
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
		table.Editions.AllColumns,
	).FROM(
		table.CollectableInstances.
			INNER_JOIN(collectable, table.CollectableInstances.CollectableID.EQ(collectable.ID)).
			INNER_JOIN(creator, collectable.CreatorID.EQ(creator.ID)).
			INNER_JOIN(owner, table.CollectableInstances.OwnerID.EQ(owner.ID)).
			INNER_JOIN(table.Editions, table.CollectableInstances.EditionID.EQ(table.Editions.ID)),
	)

	if options.ByOwner != 0 {
		stmt = stmt.WHERE(owner.ID.EQ(postgres.Int64(options.ByOwner)))
	}
	if options.ByCollectable != 0 {
		stmt = stmt.WHERE(collectable.ID.EQ(postgres.Int64(options.ByCollectable)))
	}
	if options.ByID != 0 {
		stmt = stmt.WHERE(table.CollectableInstances.ID.EQ(postgres.Int64(options.ByID)))
	}
	if !options.GetDeleted {
		stmt = stmt.WHERE(table.CollectableInstances.DeletedAt.IS_NULL())
	}

	stmt.ORDER_BY(table.CollectableInstances.CreatedAt.DESC())

	dest := []CollectableInstance{}
	err := stmt.QueryContext(ctx, db.DB, &dest)
	if err != nil {
		return nil, err
	}

	return dest, nil
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

	if !options.GetDeleted {
		stmt = stmt.WHERE(collectable.DeletedAt.IS_NULL())
	}
	if !options.GetUnapproved {
		stmt = stmt.WHERE(collectable.ApprovedAt.IS_NOT_NULL())
	}
	if options.Name != "" {
		stmt = stmt.WHERE(collectable.Name.EQ(postgres.String(options.Name)))
	}

	dest := Collectable{}

	err := stmt.QueryContext(ctx, db.DB, &dest)
	if err != nil {
		return nil, err
	}

	return &dest, nil
}

type GetCollectablesOptions struct {
	Collection    int64
	Creator       int64
	Rarity        string
	GetDeleted    bool
	GetUnapproved bool
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

	where := []postgres.BoolExpression{}
	if options.Collection != 0 {
		where = append(where, collectable.CollectionID.EQ(postgres.Int64(options.Collection)))
	}
	if options.Creator != 0 {
		where = append(where, creator.ID.EQ(postgres.Int64(options.Creator)))
	}
	if options.Rarity != "" {
		stmt = stmt.WHERE(collectable.Rarity.EQ(postgres.String(options.Rarity)))
	}
	if !options.GetDeleted {
		where = append(where, collectable.DeletedAt.IS_NULL())
	}
	if !options.GetUnapproved {
		where = append(where, collectable.ApprovedAt.IS_NOT_NULL())
	}
	if len(where) > 0 {
		stmt = stmt.WHERE(postgres.AND(where...))
	}

	stmt.ORDER_BY(collectable.CreatedAt.ASC())

	dest := []*Collectable{}

	err := stmt.QueryContext(ctx, db.DB, &dest)
	if err != nil {
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

		if options.ID != 0 {
			stmt = stmt.WHERE(table.Users.ID.EQ(postgres.Int64(options.ID)))
		}
		if options.TwitchID != "" {
			stmt = stmt.WHERE(table.Users.TwitchID.EQ(postgres.String(options.TwitchID)))
		}
		if options.Username != "" {
			stmt = stmt.WHERE(table.Users.Name.EQ(postgres.String(options.Username)))
		}
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

type GetEquippedOptions struct {
	UserID   int64
	TwitchID int64
	UserName string
}

func (db *PostgresDB) GetEquipped(ctx context.Context, options GetEquippedOptions) (*CollectableInstance, error) {
	return nil, nil
}

type SetEquippedOptions = GetEquippedOptions

func (db *PostgresDB) SetEquipped(ctx context.Context, collectableInstanceID int64, options SetEquippedOptions) (*CollectableInstance, error) {
	return nil, nil
}

func (db *PostgresDB) CreateCollectable(ctx context.Context, collectable Collectable) (*Collectable, error) {
	return nil, nil
}

func (db *PostgresDB) UpdateCollectable(ctx context.Context, collectable Collectable) (*Collectable, error) {
	return nil, nil
}

func (db *PostgresDB) DeleteCollectable(ctx context.Context, collectable Collectable) (*Collectable, error) {
	return nil, nil
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
	return nil, nil
}

func (db *PostgresDB) SetWeights(ctx context.Context, collectionID int64, weights []*PullWeight) ([]*PullWeight, error) {
	return nil, nil
}
