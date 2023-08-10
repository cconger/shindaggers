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

func (m *MockDB) GetKnife(ctx context.Context, knifeID int) (*Knife, error) {
	res := &Knife{}
	faker.FakeData(&res)

	return res, nil
}

func (m *MockDB) GetKnivesForUsername(ctx context.Context, username string) ([]*Knife, error) {
	res := make([]*Knife, 6)
	faker.FakeData(&res)

	return res, nil
}

func (m *MockDB) GetUsers(ctx context.Context, substr string) ([]*User, error) {
	return nil, nil
}

func (m *MockDB) GetUserByID(ctx context.Context, id int) (*User, error) {
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

func (m *MockDB) GetEquippedKnifeForUser(ctx context.Context, userID int) (*Knife, error) {
	res := &Knife{}
	faker.FakeData(&res)
	return res, nil
}

func (m *MockDB) EquipKnifeForUser(ctx context.Context, userID int, knifeID int) error {
	return ErrNoWrites
}

func (m *MockDB) CreateUser(ctx context.Context, user *User) (*User, error) {
	return nil, ErrNoWrites
}

func (m *MockDB) PullKnife(ctx context.Context, userID int, knifename string, subscriber bool, verified bool, edition_id int) (*Knife, error) {
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

func (m *MockDB) GetKnifeType(ctx context.Context, id int, getDeleted bool, getUnapproved bool) (*KnifeType, error) {
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

func (m *MockDB) CreateImageUpload(ctx context.Context, id int64, authorID int, path string, uploadname string) error {
	return ErrNoWrites
}

func (m *MockDB) GetWeights(ctx context.Context) ([]*PullWeight, error) {
	return []*PullWeight{}, nil
}

func (m *MockDB) SetWeights(ctx context.Context, weights []*PullWeight) ([]*PullWeight, error) {
	return nil, ErrNoWrites
}

func (m *MockDB) IssueCollectable(ctx context.Context, collectableID int, userID int, subscriber bool, verified bool, editionID int, source string) (*Knife, error) {
	return nil, ErrNoWrites
}

func (m *MockDB) ApproveKnifeType(ctx context.Context, id int, userID int) (*KnifeType, error) {
	return nil, ErrNoWrites
}
