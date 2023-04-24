package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateAd(t *testing.T) {
	client := getTestClient()

	userResp, err := client.createUser(123, "nickname", "example@mail.com")
	assert.NoError(t, err)
	assert.Equal(t, userResp.Data.ID, int64(123))
	assert.Equal(t, userResp.Data.Nickname, "nickname")
	assert.Equal(t, userResp.Data.Email, "example@mail.com")

	userResp, err = client.createUser(123, "cat", "cat@mail.com")
	assert.ErrorIs(t, err, ErrBadRequest)

	adResp, err := client.createAd(123, "hello", "world")
	assert.NoError(t, err)
	assert.Zero(t, adResp.Data.ID)
	assert.Equal(t, adResp.Data.Title, "hello")
	assert.Equal(t, adResp.Data.Text, "world")
	assert.Equal(t, adResp.Data.AuthorID, int64(123))
	assert.False(t, adResp.Data.Published)
}

func TestChangeAdStatus(t *testing.T) {
	client := getTestClient()

	_, _ = client.createUser(123, "nickname", "example@mail.com")

	response, err := client.createAd(123, "hello", "world")
	assert.NoError(t, err)

	response, err = client.changeAdStatus(123, response.Data.ID, true)
	assert.NoError(t, err)
	assert.True(t, response.Data.Published)

	response, err = client.changeAdStatus(123, response.Data.ID, false)
	assert.NoError(t, err)
	assert.False(t, response.Data.Published)

	response, err = client.changeAdStatus(123, response.Data.ID, false)
	assert.NoError(t, err)
	assert.False(t, response.Data.Published)
}

func TestUpdateAd(t *testing.T) {
	client := getTestClient()

	_, _ = client.createUser(123, "nickname", "example@mail.com")

	response, err := client.createAd(123, "hello", "world")
	assert.NoError(t, err)

	response, err = client.updateAd(123, response.Data.ID, "привет", "мир")
	assert.NoError(t, err)
	assert.Equal(t, response.Data.Title, "привет")
	assert.Equal(t, response.Data.Text, "мир")
}

func TestListAds(t *testing.T) {
	client := getTestClient()

	_, _ = client.createUser(123, "nickname", "example@mail.com")

	response, err := client.createAd(123, "hello", "world")
	assert.NoError(t, err)

	publishedAd, err := client.changeAdStatus(123, response.Data.ID, true)
	assert.NoError(t, err)

	_, err = client.createAd(123, "best cat", "not for sale")
	assert.NoError(t, err)

	ads, err := client.listAdsBasic()
	assert.NoError(t, err)
	assert.Len(t, ads.Data, 1)
	assert.Equal(t, ads.Data[0].ID, publishedAd.Data.ID)
	assert.Equal(t, ads.Data[0].Title, publishedAd.Data.Title)
	assert.Equal(t, ads.Data[0].Text, publishedAd.Data.Text)
	assert.Equal(t, ads.Data[0].AuthorID, publishedAd.Data.AuthorID)
	assert.True(t, ads.Data[0].Published)
}
