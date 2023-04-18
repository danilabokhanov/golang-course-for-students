package adfilter

import (
	"homework8/internal/app"
	"sync"
)

func New() app.Filter {
	return &BasicFilter{mx: &sync.RWMutex{}} // TODO: реализовать
}
