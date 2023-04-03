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
	Edition    string
	ObtainedAt time.Time
}

type KnifeType struct {
	ID        int
	Name      string
	ImageName string
}

type User struct {
	ID        int
	Name      string
	TwitchID  string
	CreatedAt time.Time
}

type KnifeDB interface {
	GetKnife(ctx context.Context, knifeID int) (*Knife, error)
	GetKnivesForUsername(ctx context.Context, username string) ([]*Knife, error)
	GetUser(ctx context.Context, username string) (*User, error)

	CreateUser(ctx context.Context, name string) (*User, error)
	// GetKnifeByName(ctx context.Context, knifename string) (*KnifeType, error)
	PullKnife(ctx context.Context, username string, knifename string) (*Knife, error)

	Close(context.Context) error
}
