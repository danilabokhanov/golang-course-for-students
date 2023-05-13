package customer

import (
	"homework10/internal/app"
	"homework10/internal/user"
	"sync"
)

func New() app.Users {
	return &BasicCustomer{mx: &sync.RWMutex{}, mp: map[int64]user.User{}} // TODO: реализовать
}
