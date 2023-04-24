package adrepo

import (
	"context"
	"homework9/internal/adpattern"
	"homework9/internal/ads"
	"homework9/internal/app"
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
	sort.SliceStable(res, func(i, j int) bool {
		return res[i].CreationDate.Before(res[j].CreationDate)
	})
	return res, nil
}

func (d *SliceRepo) Delete(ctx context.Context, adID int64) error {
	d.mx.Lock()
	defer d.mx.Unlock()
	delete(d.mp, adID)
	return nil
}

func (d *SliceRepo) DeleteByAuthor(ctx context.Context, userID int64) error {
	d.mx.Lock()
	defer d.mx.Unlock()

	keysToDelete := []int64{}
	for key, val := range d.mp {
		if val.AuthorID == userID {
			keysToDelete = append(keysToDelete, key)
		}
	}

	for _, key := range keysToDelete {
		delete(d.mp, key)
	}
	return nil
}
