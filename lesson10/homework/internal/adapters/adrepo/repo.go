package adrepo

import (
	"homework10/internal/ads"
	"homework10/internal/app"
	"sync"
)

func New() app.Repository {
	return &MapRepo{mx: &sync.RWMutex{}, mp: map[int64]ads.Ad{}} // TODO: реализовать
}
