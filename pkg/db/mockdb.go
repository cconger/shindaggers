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

func (m *MockDB) GetKnifeType(ctx context.Context, id int, getDeleted bool) (*KnifeType, error) {
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
