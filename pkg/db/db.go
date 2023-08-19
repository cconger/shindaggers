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
	Deleted    bool
}

type KnifeType struct {
	ID         int
	Name       string
	Author     string
	AuthorID   int
	Rarity     string
	ImageName  string
	Deleted    bool
	Approved   bool
	ApprovedAt time.Time
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

type PullWeight struct {
	CommunityID int64
	Rarity      string
	Weight      int
	UpdatedAt   time.Time
}

type CombatReport struct {
	ID           int64
	Participants []int64
	Knives       []int64
	Outcomes     []int
	Event        string
	CreatedAt    time.Time
}

type KnifeDB interface {
	GetLatestPulls(ctx context.Context) ([]*Knife, error)
	GetKnife(ctx context.Context, knifeID int) (*Knife, error)
	GetKnivesForUsername(ctx context.Context, username string) ([]*Knife, error)

	GetUsers(ctx context.Context, substr string) ([]*User, error)
	GetUserByID(ctx context.Context, id int) (*User, error)
	GetUserByTwitchID(ctx context.Context, id string) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)

	EquipKnifeForUser(ctx context.Context, userID int, KnifeID int) error
	GetEquippedKnifeForUser(ctx context.Context, id int) (*Knife, error)

	CreateUser(ctx context.Context, user *User) (*User, error)

	PullKnife(ctx context.Context, userID int, knifename string, subscriber bool, verified bool, edition_id int) (*Knife, error)
	CreateKnifeType(ctx context.Context, knife *KnifeType) (*KnifeType, error)
	CreateEdition(ctx context.Context, edition *Edition) (*Edition, error)

	GetCollection(ctx context.Context, getDeleted bool) ([]*KnifeType, error)
	GetKnifeType(ctx context.Context, id int, getDeleted bool, getUnapproved bool) (*KnifeType, error)
	GetKnifeTypesByRarity(ctx context.Context, rarity string) ([]*KnifeType, error)

	GetPendingKnives(ctx context.Context) ([]*KnifeType, error)
	ApproveKnifeType(ctx context.Context, id int, userID int) (*KnifeType, error)

	UpdateKnifeType(ctx context.Context, knife *KnifeType) (*KnifeType, error)
	DeleteKnifeType(ctx context.Context, knife *KnifeType) error

	GetEditions(ctx context.Context) ([]*Edition, error)

	// Pull Weights
	GetWeights(ctx context.Context) ([]*PullWeight, error)
	SetWeights(ctx context.Context, weights []*PullWeight) ([]*PullWeight, error)
	IssueCollectable(ctx context.Context, collectableID int, userID int, subscriber bool, verified bool, editionID int, source string) (*Knife, error)

	// ImageUpload Log
	CreateImageUpload(ctx context.Context, id int64, authorID int, path string, uploadname string) error

	// Twitch Auth
	GetAuth(ctx context.Context, token []byte) (*UserAuth, error)
	SaveAuth(ctx context.Context, auth *UserAuth) (*UserAuth, error)

	// Combat Reports
	GetCombatReport(ctx context.Context, id int64) (*CombatReport, error)
	CreateCombatReport(ctx context.Context, report *CombatReport) (*CombatReport, error)

	Close(context.Context) error
}
