package app

import (
	"context"
	"fmt"
	"github.com/danilabokhanov/strintvalidator"
	"homework6/internal/ads"
)

type App interface {
	// TODO: реализовать
	CreateAd(ctx context.Context, title string, text string, userID int64) (ads.Ad, error)
	ChangeAdStatus(ctx context.Context, adID int64, UserID int64, published bool) (ads.Ad, error)
	UpdateAd(ctx context.Context, adID int64, UserID int64, title string, text string) (ads.Ad, error)
}

type Repository interface {
	// TODO: реализовать
	Find(ctx context.Context, adID int64) (ads.Ad, bool)
	Add(ctx context.Context, title string, text string, userID int64) (int64, error)
	SetTitle(ctx context.Context, adID int64, title string) error
	SetText(ctx context.Context, adID int64, text string) error
	SetStatus(ctx context.Context, adID int64, status bool) error
}

type SimpleApp struct {
	repository Repository
}

func NewApp(repo Repository) App {
	return SimpleApp{repository: repo}
	// TODO: реализовать
}

var ErrWrongFormat = fmt.Errorf("wrong format")
var ErrWrongKey = fmt.Errorf("wrong adID")
var ErrNoAccess = fmt.Errorf("permission denied")

func (d SimpleApp) Validate(ctx context.Context, adID int64) (ads.Ad, error) {
	ad, _ := d.repository.Find(ctx, adID)
	if e := strintvalidator.Validate(ad); e != nil {
		return ads.Ad{}, ErrWrongFormat
	}
	return ad, nil
}

func (d SimpleApp) CreateAd(ctx context.Context, title string, text string, userID int64) (ads.Ad, error) {
	if e := strintvalidator.Validate(ads.Ad{Title: title, Text: text}); e != nil {
		return ads.Ad{}, ErrWrongFormat
	}
	adID, err := d.repository.Add(ctx, title, text, userID)
	if err != nil {
		return ads.Ad{}, err
	}
	ad, _ := d.repository.Find(ctx, adID)
	return ad, nil
}

func (d SimpleApp) ChangeAdStatus(ctx context.Context, adID int64, UserID int64, published bool) (ads.Ad, error) {
	ad, isFound := d.repository.Find(ctx, adID)
	if !isFound {
		return ads.Ad{}, ErrWrongKey
	}
	if ad.AuthorID != UserID {
		return ads.Ad{}, ErrNoAccess
	}
	err := d.repository.SetStatus(ctx, adID, published)
	if err != nil {
		return ads.Ad{}, err
	}
	ad.Published = published
	return ad, nil
}

func (d SimpleApp) UpdateAd(ctx context.Context, adID int64, UserID int64, title string, text string) (ads.Ad, error) {
	if e := strintvalidator.Validate(ads.Ad{Title: title, Text: text}); e != nil {
		return ads.Ad{}, ErrWrongFormat
	}
	ad, isFound := d.repository.Find(ctx, adID)
	if !isFound {
		return ads.Ad{}, ErrWrongKey
	}
	if ad.AuthorID != UserID {
		return ads.Ad{}, ErrNoAccess
	}
	err := d.repository.SetText(ctx, adID, text)
	if err != nil {
		return ads.Ad{}, err
	}
	err = d.repository.SetTitle(ctx, adID, title)
	if err != nil {
		return ads.Ad{}, err
	}
	ad.Title = title
	ad.Text = text
	return ad, nil
}
