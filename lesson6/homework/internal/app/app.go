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
	Find(ctx context.Context, adID int64) (ads.Ad, error)
	Add(ctx context.Context, title string, text string, userID int64) (int64, error)
	SetTitle(ctx context.Context, adID, UserID int64, title string) error
	SetText(ctx context.Context, adID, UserID int64, text string) error
	SetStatus(ctx context.Context, adID, UserID int64, status bool) error
}

type SimpleApp struct {
	repository Repository
}

func NewApp(repo Repository) App {
	return SimpleApp{repository: repo}
	// TODO: реализовать
}

var ErrWrongFormat = fmt.Errorf("wrong format")

func (d SimpleApp) FindAndValidate(ctx context.Context, adID int64) (ads.Ad, error) {
	ad, err := d.repository.Find(ctx, adID)
	if err != nil {
		return ads.Ad{}, err
	}
	if e := strintvalidator.Validate(ad); e != nil {
		return ads.Ad{}, ErrWrongFormat
	}
	return ad, nil
}

func (d SimpleApp) CreateAd(ctx context.Context, title string, text string, userID int64) (ads.Ad, error) {
	adID, err := d.repository.Add(ctx, title, text, userID)
	if err != nil {
		return ads.Ad{}, err
	}
	return d.FindAndValidate(ctx, adID)
}

func (d SimpleApp) ChangeAdStatus(ctx context.Context, adID int64, UserID int64, published bool) (ads.Ad, error) {
	err := d.repository.SetStatus(ctx, adID, UserID, published)
	if err != nil {
		return ads.Ad{}, err
	}
	return d.FindAndValidate(ctx, adID)
}

func (d SimpleApp) UpdateAd(ctx context.Context, adID int64, UserID int64, title string, text string) (ads.Ad, error) {
	err := d.repository.SetText(ctx, adID, UserID, text)
	if err != nil {
		return ads.Ad{}, err
	}
	err = d.repository.SetTitle(ctx, adID, UserID, title)
	if err != nil {
		return ads.Ad{}, err
	}
	return d.FindAndValidate(ctx, adID)
}
