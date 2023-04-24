package customer

import (
	"homework9/internal/app"
	"homework9/internal/user"
	"sync"
)

func New() app.Users {
	return &BasicCustomer{mx: &sync.RWMutex{}, mp: map[int64]user.User{}} // TODO: реализовать
}
