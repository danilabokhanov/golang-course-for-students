package adrepo

import (
	"homework8/internal/ads"
	"homework8/internal/app"
	"sync"
)

func New() app.Repository {
	return &SliceRepo{mx: &sync.RWMutex{}, mp: map[int64]ads.Ad{}} // TODO: реализовать
}
