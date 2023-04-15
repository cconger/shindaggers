package db

import (
	"context"
	"time"
)

type Knife struct {
	ID         int
	InstanceID int
	Name       string
	Author     string
	AuthorID   int
	Owner      string
	OwnerID    int
	Rarity     string
	ImageName  string
	Verified   bool
	Subscriber bool
	Edition    string
	ObtainedAt time.Time
}

type KnifeType struct {
	ID        int
	Name      string
	Author    string
	AuthorID  int
	Rarity    string
	ImageName string
}

type User struct {
	ID         int
	Name       string
	LookupName string
	TwitchID   string
	Admin      bool
	CreatedAt  time.Time
}

type UserAuth struct {
	UserID       int
	Token        []byte
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	UpdatedAt    time.Time
}

type Edition struct {
	ID        int
	Name      string
	UpdatedAt time.Time
}

type KnifeDB interface {
	GetLatestPulls(ctx context.Context) ([]*Knife, error)
	GetKnife(ctx context.Context, knifeID int) (*Knife, error)
	GetKnivesForUsername(ctx context.Context, username string) ([]*Knife, error)

	GetUserByID(ctx context.Context, id int) (*User, error)
	GetUserByTwitchID(ctx context.Context, id string) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)

	CreateUser(ctx context.Context, user *User) (*User, error)

	PullKnife(ctx context.Context, userID int, knifename string, subscriber bool, verified bool, edition_id int) (*Knife, error)
	CreateKnifeType(ctx context.Context, knife *KnifeType) (*KnifeType, error)
	CreateEdition(ctx context.Context, edition *Edition) (*Edition, error)

	GetCollection(ctx context.Context) ([]*KnifeType, error)
	GetKnifeType(ctx context.Context, id int) (*KnifeType, error)

	UpdateKnifeType(ctx context.Context, knife *KnifeType) (*KnifeType, error)
	DeleteKnifeType(ctx context.Context, knife *KnifeType) error

	GetEditions(ctx context.Context) ([]*Edition, error)

	// Twitch Auth
	GetAuth(ctx context.Context, token []byte) (*UserAuth, error)
	SaveAuth(ctx context.Context, auth *UserAuth) (*UserAuth, error)

	Close(context.Context) error
}
