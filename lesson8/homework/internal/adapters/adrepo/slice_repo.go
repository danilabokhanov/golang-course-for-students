package adrepo

import (
	"context"
	"homework8/internal/adpattern"
	"homework8/internal/ads"
	"homework8/internal/app"
	"sort"
	"strings"
	"sync"
	"time"
)

type SliceRepo struct {
	mx    *sync.RWMutex
	mp    map[int64]ads.Ad
	curID int64
}

func (d *SliceRepo) Find(ctx context.Context, adID int64) (ads.Ad, bool) {
	d.mx.RLock()
	defer d.mx.RUnlock()
	if _, ok := d.mp[adID]; !ok {
		return ads.Ad{}, false
	}
	return d.mp[adID], true
}

func (d *SliceRepo) Add(ctx context.Context, title string, text string, userID int64) (int64, error) {
	d.mx.Lock()
	defer d.mx.Unlock()

	for {
		if _, ok := d.mp[d.curID]; !ok {
			break
		}
		d.curID++
	}
	d.mp[d.curID] = ads.Ad{ID: d.curID, Title: title, Text: text, AuthorID: userID,
		Published: false, CreationDate: time.Now().UTC(), UpdateDate: time.Now().UTC()}
	return d.curID, nil
}

func (d *SliceRepo) SetTitle(ctx context.Context, adID int64, title string) error {
	d.mx.Lock()
	defer d.mx.Unlock()
	cur := d.mp[adID]
	cur.Title = title
	cur.UpdateDate = time.Now().UTC()
	d.mp[adID] = cur
	return nil
}

func (d *SliceRepo) SetText(ctx context.Context, adID int64, text string) error {
	d.mx.Lock()
	defer d.mx.Unlock()
	cur := d.mp[adID]
	cur.Text = text
	cur.UpdateDate = time.Now().UTC()
	d.mp[adID] = cur
	return nil
}

func (d *SliceRepo) SetStatus(ctx context.Context, adID int64, status bool) error {
	d.mx.Lock()
	defer d.mx.Unlock()
	cur := d.mp[adID]
	cur.Published = status
	cur.UpdateDate = time.Now().UTC()
	d.mp[adID] = cur
	return nil
}

func (d *SliceRepo) GetAllByTemplate(ctx context.Context, adp adpattern.AdPattern) ([]ads.Ad, error) {
	d.mx.RLock()
	defer d.mx.RUnlock()
	res := []ads.Ad{}
	for _, ad := range d.mp {
		if app.CheckAd(ad, adp) {
			res = append(res, ad)
		}
	}
	sort.SliceStable(res, func(i, j int) bool {
		return res[i].CreationDate.Before(res[j].CreationDate)
	})
	return res, nil
}

func (d *SliceRepo) GetByTitle(ctx context.Context, title string) ([]ads.Ad, error) {
	d.mx.RLock()
	defer d.mx.RUnlock()
	res := []ads.Ad{}
	for _, ad := range d.mp {
		if strings.HasPrefix(ad.Title, title) {
			res = append(res, ad)
		}
	}
	return res, nil
}
