package adrepo

import (
	"homework9/internal/ads"
	"homework9/internal/app"
	"sync"
)

func New() app.Repository {
	return &SliceRepo{mx: &sync.RWMutex{}, mp: map[int64]ads.Ad{}} // TODO: реализовать
}
