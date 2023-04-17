package customer

import (
	"homework8/internal/app"
	"homework8/internal/user"
	"sync"
)

func New() app.Users {
	return &BasicCustomer{mx: &sync.RWMutex{}, mp: map[int64]user.User{}} // TODO: реализовать
}
