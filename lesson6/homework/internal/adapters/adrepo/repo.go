package adrepo

import (
	"homework6/internal/app"
)

func New() app.Repository {
	return &app.SliceRepo{} // TODO: реализовать
}
