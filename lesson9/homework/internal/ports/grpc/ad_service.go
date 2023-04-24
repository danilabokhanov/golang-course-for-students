package grpc

import "homework9/internal/app"

type AdService struct {
	a app.App
}

func NewService(a app.App) AdService {
	return AdService{a}
}
