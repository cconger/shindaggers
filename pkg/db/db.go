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
	Edition    string
	ObtainedAt time.Time
}

type User struct {
	ID        int
	Name      string
	TwitchID  string
	CreatedAt time.Time
}

type KnifeDB interface {
	GetKnife(ctx context.Context, knifeID int) (*Knife, error)
	GetKnivesForUser(ctx context.Context, userID int) ([]*Knife, error)
	Close(context.Context) error
}
