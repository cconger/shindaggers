package db

import (
	"context"
	"time"
)

type Knife struct {
	ID         int64
	InstanceID int64
	Name       string
	Author     string
	AuthorID   int64
	Owner      string
	OwnerID    int64
	Rarity     string
	ImageName  string
	Verified   bool
	Subscriber bool
	Edition    string
	ObtainedAt time.Time
	Deleted    bool
}

type KnifeType struct {
	ID         int64
	Name       string
	Author     string
	AuthorID   int64
	Rarity     string
	ImageName  string
	Deleted    bool
	Approved   bool
	ApprovedAt time.Time
}

type User struct {
	ID         int64
	Name       string
	LookupName string
	TwitchID   string
	Admin      bool
	CreatedAt  time.Time
}

type UserAuth struct {
	UserID       int64
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

type CombatOutcome struct {
	FightID       int64
	UserID        int64
	CollectableID int64
	Outcome       string
	EventID       *int64
}

type Event struct {
	ID          int64
	Name        string
	Description string
	CreatedAt   time.Time
}

type CombatStats = map[string]int

type KnifeDB interface {
	GetLP(ctx context.Context) ([]JKnifeInstance, error)
	GetLatestPulls(ctx context.Context) ([]*Knife, error)
	GetLatestPullsSince(ctx context.Context, since time.Time) ([]*Knife, error)
	GetKnife(ctx context.Context, knifeID int64) (*Knife, error)
	GetKnives(ctx context.Context, knifeID ...int64) ([]*Knife, error)
	GetKnivesForUser(ctx context.Context, userID int64) ([]*Knife, error)

	GetUsers(ctx context.Context, substr string) ([]*User, error)
	GetUserByID(ctx context.Context, id int64) (*User, error)
	GetUsersByID(ctx context.Context, ids ...int64) ([]*User, error)
	GetUserByTwitchID(ctx context.Context, id string) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)

	EquipKnifeForUser(ctx context.Context, userID int64, KnifeID int64) error
	GetEquippedKnifeForUser(ctx context.Context, id int64) (*Knife, error)

	CreateUser(ctx context.Context, user *User) (*User, error)
	UpdateUser(ctx context.Context, user *User) (*User, error)
	CreateKnifeType(ctx context.Context, knife *KnifeType) (*KnifeType, error)
	CreateEdition(ctx context.Context, edition *Edition) (*Edition, error)
	IssueCollectable(ctx context.Context, knife *Knife, source string) (*Knife, error)

	GetCollection(ctx context.Context, getDeleted bool) ([]*KnifeType, error)
	GetKnifeType(ctx context.Context, id int64, getDeleted bool, getUnapproved bool) (*KnifeType, error)
	GetKnifeTypeByName(ctx context.Context, name string) (*KnifeType, error)
	GetKnifeTypesByRarity(ctx context.Context, rarity string) ([]*KnifeType, error)

	GetPendingKnives(ctx context.Context) ([]*KnifeType, error)
	ApproveKnifeType(ctx context.Context, id int64, userID int64) (*KnifeType, error)

	UpdateKnifeType(ctx context.Context, knife *KnifeType) (*KnifeType, error)
	DeleteKnifeType(ctx context.Context, knife *KnifeType) error

	GetEditions(ctx context.Context) ([]*Edition, error)

	// Pull Weights
	GetWeights(ctx context.Context) ([]*PullWeight, error)
	SetWeights(ctx context.Context, weights []*PullWeight) ([]*PullWeight, error)

	// ImageUpload Log
	CreateImageUpload(ctx context.Context, id int64, authorID int64, path string, uploadname string) error

	// Twitch Auth
	GetAuth(ctx context.Context, token []byte) (*UserAuth, error)
	SaveAuth(ctx context.Context, auth *UserAuth) (*UserAuth, error)

	// Combat Reports
	GetCombatReport(ctx context.Context, id int64) (*CombatReport, error)
	CreateCombatReport(ctx context.Context, report *CombatReport) (*CombatReport, error)
	GetCombatReportsForEvent(ctx context.Context, event string) ([]*CombatReport, error)

	// Get CombatStats
	GetCombatStatsForUser(ctx context.Context, userID int64) (CombatStats, error)
	GetCombatStatsForKnife(ctx context.Context, knifeID int64) (CombatStats, error)

	// Events
	// CreateEvent(ctx context.Context, event *Event) (*Event, error)
	// GetEvent(ctx context.Context, id int64) (*Event, error)

	Close(context.Context) error
}
