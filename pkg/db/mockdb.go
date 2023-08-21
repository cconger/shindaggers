package db

import (
	"context"
	"fmt"

	"github.com/go-faker/faker/v4"
)

type MockDB struct{}

var ErrNoWrites = fmt.Errorf("write operations not supported on dev db")

// Compile time enforce that our mock implements the interface
var _ KnifeDB = &MockDB{}

func (m *MockDB) GetLatestPulls(ctx context.Context) ([]*Knife, error) {
	res := make([]*Knife, 10)
	faker.FakeData(&res)

	fmt.Println(res)

	return res, nil
}

func (m *MockDB) GetKnife(ctx context.Context, knifeID int64) (*Knife, error) {
	res := &Knife{}
	faker.FakeData(&res)

	return res, nil
}

func (m *MockDB) GetKnivesForUser(ctx context.Context, userID int64) ([]*Knife, error) {
	res := make([]*Knife, 6)
	faker.FakeData(&res)

	return res, nil
}

func (m *MockDB) GetUsers(ctx context.Context, substr string) ([]*User, error) {
	return nil, nil
}

func (m *MockDB) GetUserByID(ctx context.Context, id int64) (*User, error) {
	res := &User{}
	faker.FakeData(&res)
	return res, nil
}

func (m *MockDB) GetUserByTwitchID(ctx context.Context, id string) (*User, error) {
	res := &User{}
	faker.FakeData(&res)
	return res, nil
}

func (m *MockDB) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	res := &User{}
	faker.FakeData(&res)
	return res, nil
}

func (m *MockDB) GetEquippedKnifeForUser(ctx context.Context, userID int64) (*Knife, error) {
	res := &Knife{}
	faker.FakeData(&res)
	return res, nil
}

func (m *MockDB) EquipKnifeForUser(ctx context.Context, userID int64, knifeID int64) error {
	return ErrNoWrites
}

func (m *MockDB) CreateUser(ctx context.Context, user *User) (*User, error) {
	return nil, ErrNoWrites
}

func (m *MockDB) CreateKnifeType(ctx context.Context, knife *KnifeType) (*KnifeType, error) {
	return nil, ErrNoWrites
}

func (m *MockDB) CreateEdition(ctx context.Context, edition *Edition) (*Edition, error) {
	return nil, ErrNoWrites
}

func (m *MockDB) GetCollection(ctx context.Context, getDeleted bool) ([]*KnifeType, error) {
	return nil, nil
}

func (m *MockDB) GetPendingKnives(ctx context.Context) ([]*KnifeType, error) {
	return nil, nil
}

func (m *MockDB) GetKnifeTypesByRarity(ctx context.Context, rarity string) ([]*KnifeType, error) {
	return nil, nil
}

func (m *MockDB) GetKnifeType(ctx context.Context, id int64, getDeleted bool, getUnapproved bool) (*KnifeType, error) {
	res := &KnifeType{}
	faker.FakeData(&res)
	return res, nil
}

func (m *MockDB) GetKnifeTypeByName(ctx context.Context, name string) (*KnifeType, error) {
	res := &KnifeType{}
	faker.FakeData(&res)
	return res, nil
}

func (m *MockDB) UpdateKnifeType(ctx context.Context, knife *KnifeType) (*KnifeType, error) {
	return nil, ErrNoWrites
}

func (m *MockDB) DeleteKnifeType(ctx context.Context, knife *KnifeType) error {
	return ErrNoWrites
}

func (m *MockDB) GetEditions(ctx context.Context) ([]*Edition, error) {
	edition := make([]*Edition, 3)
	faker.FakeData(&edition)
	return edition, nil
}

func (m *MockDB) GetAuth(ctx context.Context, token []byte) (*UserAuth, error) {
	return nil, ErrNoWrites
}

func (m *MockDB) SaveAuth(ctx context.Context, auth *UserAuth) (*UserAuth, error) {
	return nil, ErrNoWrites
}

func (m *MockDB) Close(context.Context) error {
	return nil
}

func (m *MockDB) CreateImageUpload(ctx context.Context, id int64, authorID int64, path string, uploadname string) error {
	return ErrNoWrites
}

func (m *MockDB) GetWeights(ctx context.Context) ([]*PullWeight, error) {
	return []*PullWeight{}, nil
}

func (m *MockDB) SetWeights(ctx context.Context, weights []*PullWeight) ([]*PullWeight, error) {
	return nil, ErrNoWrites
}

func (m *MockDB) IssueCollectable(ctx context.Context, k *Knife, source string) (*Knife, error) {
	return nil, ErrNoWrites
}

func (m *MockDB) ApproveKnifeType(ctx context.Context, id int64, userID int64) (*KnifeType, error) {
	return nil, ErrNoWrites
}

func (m *MockDB) GetCombatReport(ctx context.Context, id int64) (*CombatReport, error) {
	return nil, ErrNotFound
}

func (m *MockDB) CreateCombatReport(ctx context.Context, report *CombatReport) (*CombatReport, error) {
	return nil, ErrNoWrites
}

func (m *MockDB) GetCombatStatsForUser(ctx context.Context, userID int64) (CombatStats, error) {
	return nil, nil
}

func (m *MockDB) GetCombatStatsForKnife(ctx context.Context, userID int64) (CombatStats, error) {
	return nil, nil
}
