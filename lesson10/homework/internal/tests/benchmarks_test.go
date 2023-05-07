package tests

import (
	"context"
	"fmt"
	"homework10/internal/adapters/adrepo"
	"homework10/internal/adapters/customer"
	"strconv"
	"testing"
)

const (
	Users = 50000
	Ads   = 1000000
)

func BenchmarkMapRepo(b *testing.B) {
	ctx := context.Background()
	mapRepo := adrepo.New()
	for i := 0; i < Ads; i++ {
		_, _ = mapRepo.Add(ctx, fmt.Sprint("ad", i), "test ad", 1)
	}
}

func BenchmarkBasicCustomer(b *testing.B) {
	ctx := context.Background()
	basicCustomer := customer.New()
	for i := 0; i < Users; i++ {
		_, _ = basicCustomer.CreateByID(ctx, fmt.Sprint("user", i),
			"example"+strconv.Itoa(i)+"@mail.ru", int64(i))
	}
}
