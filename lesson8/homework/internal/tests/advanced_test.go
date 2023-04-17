package tests

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGetAdByID(t *testing.T) {
	client := getTestClient()

	client.createUser(3, "nickname", "example@mail.com")
	client.createUser(5, "cat", "cat@mail.com")
	client.createUser(7, "aba", "caba@mail.com")
	client.createAd(3, "aba", "caba")
	client.createAd(5, "foo", "bar")
	client.createAd(7, "alpha", "beta")
	response, err := client.getAdByID(1)
	assert.NoError(t, err)
	assert.Equal(t, response.Data.ID, int64(1))
	assert.Equal(t, response.Data.Title, "foo")
	assert.Equal(t, response.Data.Text, "bar")
	assert.Equal(t, response.Data.AuthorID, int64(5))
	assert.False(t, response.Data.Published)
	_, err = client.getAdByID(3)
	assert.ErrorIs(t, err, ErrBadRequest)
}

func TestCreateUser(t *testing.T) {
	client := getTestClient()

	userResp, err := client.createUser(123, "nickname", "example@mail.com")
	assert.NoError(t, err)
	assert.Equal(t, userResp.Data.ID, int64(123))
	assert.Equal(t, userResp.Data.Nickname, "nickname")
	assert.Equal(t, userResp.Data.Email, "example@mail.com")

	userResp, err = client.createUser(123, "cat", "cat@mail.com")
	assert.ErrorIs(t, err, ErrBadRequest)

	userResp, err = client.createUser(125, "aba", "caba@mail.com")
	assert.NoError(t, err)
	assert.Equal(t, userResp.Data.ID, int64(125))
	assert.Equal(t, userResp.Data.Nickname, "aba")
	assert.Equal(t, userResp.Data.Email, "caba@mail.com")
}

func TestFilterByAuthor(t *testing.T) {
	client := getTestClient()

	client.createUser(3, "nickname", "example@mail.com")
	client.createUser(5, "cat", "cat@mail.com")
	client.createUser(7, "aba", "caba@mail.com")

	a, _ := client.createAd(3, "aba", "caba")
	b, _ := client.createAd(3, "bab", "abac")
	c, _ := client.createAd(5, "foo", "bar")
	d, _ := client.createAd(7, "alpha", "beta")
	a, _ = client.changeAdStatus(3, a.Data.ID, true)
	b, _ = client.changeAdStatus(3, b.Data.ID, true)
	c, _ = client.changeAdStatus(5, c.Data.ID, true)
	d, _ = client.changeAdStatus(7, d.Data.ID, true)

	ads, err := client.listAdsAuthor(3)
	assert.NoError(t, err)
	assert.Equal(t, ads.Data[0].ID, a.Data.ID)
	assert.Equal(t, ads.Data[0].Title, a.Data.Title)
	assert.Equal(t, ads.Data[0].Text, a.Data.Text)
	assert.Equal(t, ads.Data[0].AuthorID, a.Data.AuthorID)
	assert.Equal(t, ads.Data[0].Published, a.Data.Published)

	assert.Equal(t, ads.Data[1].ID, b.Data.ID)
	assert.Equal(t, ads.Data[1].Title, b.Data.Title)
	assert.Equal(t, ads.Data[1].Text, b.Data.Text)
	assert.Equal(t, ads.Data[1].AuthorID, b.Data.AuthorID)
	assert.Equal(t, ads.Data[1].Published, b.Data.Published)
}

func TestFilterByTime(t *testing.T) {
	client := getTestClient()

	client.createUser(3, "nickname", "example@mail.com")
	client.createUser(5, "cat", "cat@mail.com")
	client.createUser(7, "aba", "caba@mail.com")

	a, _ := client.createAd(3, "aba", "caba")
	time.Sleep(2 * time.Second)
	tm := time.Now()
	lTm := tm.Add(-time.Second).UnixMicro()
	rTm := tm.Add(time.Second).UnixMicro()
	b, _ := client.createAd(3, "bab", "abac")
	c, _ := client.createAd(5, "foo", "bar")
	time.Sleep(2 * time.Second)
	d, _ := client.createAd(7, "alpha", "beta")
	a, _ = client.changeAdStatus(3, a.Data.ID, true)
	b, _ = client.changeAdStatus(3, b.Data.ID, true)
	c, _ = client.changeAdStatus(5, c.Data.ID, true)
	d, _ = client.changeAdStatus(7, d.Data.ID, true)

	ads, err := client.listAdsTime(lTm, rTm)
	assert.NoError(t, err)
	assert.Equal(t, ads.Data[0].ID, b.Data.ID)
	assert.Equal(t, ads.Data[0].Title, b.Data.Title)
	assert.Equal(t, ads.Data[0].Text, b.Data.Text)
	assert.Equal(t, ads.Data[0].AuthorID, b.Data.AuthorID)
	assert.Equal(t, ads.Data[0].Published, b.Data.Published)

	assert.Equal(t, ads.Data[1].ID, c.Data.ID)
	assert.Equal(t, ads.Data[1].Title, c.Data.Title)
	assert.Equal(t, ads.Data[1].Text, c.Data.Text)
	assert.Equal(t, ads.Data[1].AuthorID, c.Data.AuthorID)
	assert.Equal(t, ads.Data[1].Published, c.Data.Published)
}

func TestFilterByPublishedOnly(t *testing.T) {
	client := getTestClient()

	client.createUser(3, "nickname", "example@mail.com")
	client.createUser(5, "cat", "cat@mail.com")
	client.createUser(7, "aba", "caba@mail.com")

	a, _ := client.createAd(3, "aba", "caba")
	b, _ := client.createAd(3, "bab", "abac")
	c, _ := client.createAd(5, "foo", "bar")
	d, _ := client.createAd(7, "alpha", "beta")
	a, _ = client.changeAdStatus(3, a.Data.ID, false)
	b, _ = client.changeAdStatus(3, b.Data.ID, true)
	c, _ = client.changeAdStatus(5, c.Data.ID, true)
	d, _ = client.changeAdStatus(7, d.Data.ID, false)

	ads, err := client.listAdsPublishedOnly(true)
	assert.NoError(t, err)
	assert.Equal(t, ads.Data[0].ID, b.Data.ID)
	assert.Equal(t, ads.Data[0].Title, b.Data.Title)
	assert.Equal(t, ads.Data[0].Text, b.Data.Text)
	assert.Equal(t, ads.Data[0].AuthorID, b.Data.AuthorID)
	assert.Equal(t, ads.Data[0].Published, b.Data.Published)

	assert.Equal(t, ads.Data[1].ID, c.Data.ID)
	assert.Equal(t, ads.Data[1].Title, c.Data.Title)
	assert.Equal(t, ads.Data[1].Text, c.Data.Text)
	assert.Equal(t, ads.Data[1].AuthorID, c.Data.AuthorID)
	assert.Equal(t, ads.Data[1].Published, c.Data.Published)
}

func TestChangeUserInfo(t *testing.T) {
	client := getTestClient()

	client.createUser(3, "nickname", "example@mail.com")
	client.createUser(5, "cat", "cat@mail.com")
	client.createUser(7, "aba", "caba@mail.com")

	response, err := client.changeUserInfo(3, "123", "qwerty@mail.ru")
	assert.NoError(t, err)
	assert.Equal(t, response.Data.Nickname, "123")
	assert.Equal(t, response.Data.Email, "qwerty@mail.ru")

	response, err = client.changeUserInfo(1, "123", "qwerty@mail.ru")
	assert.ErrorIs(t, err, ErrBadRequest)
}

func TestGetAdsByTitle(t *testing.T) {
	client := getTestClient()

	client.createUser(3, "nickname", "example@mail.com")
	client.createUser(5, "cat", "cat@mail.com")
	client.createUser(7, "aba", "caba@mail.com")

	a, _ := client.createAd(3, "aba", "caba")
	b, _ := client.createAd(5, "aba", "abac")
	c, _ := client.createAd(7, "foo", "bar")
	client.createAd(3, "alpha", "beta")
	c, _ = client.updateAd(7, c.Data.ID, "aba", "test")

	ads, err := client.getAdsByTitle("aba")
	assert.NoError(t, err)
	assert.Equal(t, ads.Data[0].ID, a.Data.ID)
	assert.Equal(t, ads.Data[0].Title, a.Data.Title)
	assert.Equal(t, ads.Data[0].Text, a.Data.Text)
	assert.Equal(t, ads.Data[0].AuthorID, a.Data.AuthorID)
	assert.Equal(t, ads.Data[0].Published, a.Data.Published)

	assert.Equal(t, ads.Data[1].ID, b.Data.ID)
	assert.Equal(t, ads.Data[1].Title, b.Data.Title)
	assert.Equal(t, ads.Data[1].Text, b.Data.Text)
	assert.Equal(t, ads.Data[1].AuthorID, b.Data.AuthorID)
	assert.Equal(t, ads.Data[1].Published, b.Data.Published)

	assert.Equal(t, ads.Data[2].ID, c.Data.ID)
	assert.Equal(t, ads.Data[2].Title, c.Data.Title)
	assert.Equal(t, ads.Data[2].Text, c.Data.Text)
	assert.Equal(t, ads.Data[2].AuthorID, c.Data.AuthorID)
	assert.Equal(t, ads.Data[2].Published, c.Data.Published)
}

func TestGetUserByID(t *testing.T) {
	client := getTestClient()

	a, _ := client.createUser(3, "nickname", "example@mail.com")
	b, _ := client.createUser(5, "cat", "cat@mail.com")
	c, _ := client.createUser(7, "aba", "caba@mail.com")

	response, err := client.getUserByID(3)
	assert.NoError(t, err)
	assert.Equal(t, response.Data.ID, a.Data.ID)
	assert.Equal(t, response.Data.Nickname, a.Data.Nickname)
	assert.Equal(t, response.Data.Email, a.Data.Email)

	response, err = client.getUserByID(5)
	assert.NoError(t, err)
	assert.Equal(t, response.Data.ID, b.Data.ID)
	assert.Equal(t, response.Data.Nickname, b.Data.Nickname)
	assert.Equal(t, response.Data.Email, b.Data.Email)

	response, err = client.getUserByID(7)
	assert.NoError(t, err)
	assert.Equal(t, response.Data.ID, c.Data.ID)
	assert.Equal(t, response.Data.Nickname, c.Data.Nickname)
	assert.Equal(t, response.Data.Email, c.Data.Email)

	_, err = client.getUserByID(9)
	assert.ErrorIs(t, err, ErrBadRequest)
}
