package tests

import (
	"context"
	"github.com/stretchr/testify/assert"
	"homework10/internal/adapters/adfilter"
	"homework10/internal/adapters/adrepo"
	"homework10/internal/adapters/customer"
	"homework10/internal/ads"
	"homework10/internal/app"
	"testing"
)

func TestSimpleApp_InsertAdd(t *testing.T) {
	var AuthorID int64 = 3

	type SimpleAppTest struct {
		name     string
		title    string
		text     string
		userID   int64
		expected ads.Ad
	}

	simpleAppTests := [...]SimpleAppTest{
		{"Successful addition", "aba", "caba", AuthorID,
			ads.Ad{AuthorID: 3, Title: "aba", Text: "caba", Published: false}},
		{"Can't create", "cat", "qwerty", AuthorID + 1,
			ads.Ad{}},
	}

	for _, test := range simpleAppTests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			a := app.NewApp(adrepo.New(), customer.New(), adfilter.New())
			_, err := a.CreateUserByID(context.Background(), "nickname",
				"example@mail.ru", AuthorID)
			assert.NoError(t, err)
			given, _ := a.CreateAd(context.Background(), test.title, test.text, test.userID)
			assert.Equal(t, given.Title, test.expected.Title)
			assert.Equal(t, given.Text, test.expected.Text)
			assert.Equal(t, given.AuthorID, test.expected.AuthorID)
			assert.Equal(t, given.Published, test.expected.Published)
		})
	}
}

func TestSimpleApp_DeleteAdd(t *testing.T) {
	var (
		AuthorID int64 = 3
		AdID     int64 = 0
	)
	type SimpleAppTest struct {
		name     string
		id       int64
		userID   int64
		expected ads.Ad
	}

	simpleAppTests := [...]SimpleAppTest{
		{"Successful removal", AdID, AuthorID,
			ads.Ad{AuthorID: 3, ID: AdID}},
		{"Wrong Author", AdID, AuthorID + 1,
			ads.Ad{}},
		{"Wrong Id", AdID + 1, AuthorID,
			ads.Ad{}},
	}

	for _, test := range simpleAppTests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			a := app.NewApp(adrepo.New(), customer.New(), adfilter.New())
			_, err := a.CreateUserByID(context.Background(), "nickname",
				"example@mail.ru", AuthorID)
			assert.NoError(t, err)
			_, _ = a.CreateAd(context.Background(), "aba", "caba", test.userID)

			given, _ := a.DeleteAd(context.Background(), test.id, test.userID)
			assert.Equal(t, given.AuthorID, test.expected.AuthorID)
			assert.Equal(t, given.ID, test.expected.ID)
		})
	}
}
