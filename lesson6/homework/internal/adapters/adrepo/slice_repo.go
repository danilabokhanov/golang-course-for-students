package adrepo

import (
	"context"
	"homework6/internal/ads"
)

type SliceRepo struct {
	mp    []ads.Ad
	curID int64
}

func (d *SliceRepo) Find(ctx context.Context, adID int64) (ads.Ad, bool) {
	if adID >= int64(len(d.mp)) {
		return ads.Ad{}, false
	}
	return d.mp[adID], true
}

func (d *SliceRepo) Add(ctx context.Context, title string, text string, userID int64) (int64, error) {
	d.mp = append(d.mp, ads.Ad{ID: d.curID, Title: title, Text: text, AuthorID: userID, Published: false})
	d.curID++
	return d.curID - 1, nil
}

func (d *SliceRepo) SetTitle(ctx context.Context, adID, UserID int64, title string) error {
	d.mp[adID].Title = title
	return nil
}

func (d *SliceRepo) SetText(ctx context.Context, adID, UserID int64, text string) error {
	d.mp[adID].Text = text
	return nil
}

func (d *SliceRepo) SetStatus(ctx context.Context, adID, UserID int64, status bool) error {
	d.mp[adID].Published = status
	return nil
}
