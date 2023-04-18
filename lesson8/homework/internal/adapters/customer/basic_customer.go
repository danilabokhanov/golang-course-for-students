package customer

import (
	"context"
	"homework8/internal/user"
	"sync"
)

type BasicCustomer struct {
	mx    *sync.RWMutex
	mp    map[int64]user.User
	curID int64
}

func (d *BasicCustomer) Find(ctx context.Context, userID int64) (user.User, bool) {
	d.mx.Lock()
	defer d.mx.Unlock()
	if _, ok := d.mp[userID]; !ok {
		return user.User{}, false
	}
	return d.mp[userID], true
}

func (d *BasicCustomer) Add(ctx context.Context, nickname, email string) (int64, error) {
	d.mx.Lock()
	defer d.mx.Unlock()

	for {
		d.curID++
		if _, ok := d.mp[d.curID]; !ok {
			break
		}
	}
	d.mp[d.curID] = user.User{ID: d.curID, Nickname: nickname, Email: email}
	return d.curID, nil
}

func (d *BasicCustomer) ChangeInfo(ctx context.Context, userID int64, nickname, email string) error {
	d.mx.Lock()
	defer d.mx.Unlock()
	cur := d.mp[userID]
	cur.Nickname = nickname
	cur.Email = email
	d.mp[userID] = cur
	return nil
}

func (d *BasicCustomer) CreateByID(ctx context.Context, nickname string, email string, userID int64) (user.User, error) {
	d.mx.Lock()
	defer d.mx.Unlock()
	d.mp[userID] = user.User{ID: userID, Nickname: nickname, Email: email}
	return d.mp[userID], nil
}
