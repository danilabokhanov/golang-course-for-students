package tests

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"homework9/internal/adapters/adfilter"
	"homework9/internal/adapters/customer"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"homework9/internal/adapters/adrepo"
	"homework9/internal/app"
	grpcPort "homework9/internal/ports/grpc"
)

var (
	ErrorBadRequest = status.Error(codes.InvalidArgument, app.ErrWrongFormat.Error())
	ErrorForbidden  = status.Error(codes.PermissionDenied, app.ErrNoAccess.Error())
	ErrorNotFound   = status.Error(codes.NotFound, "")
)

func GetClient(t *testing.T) (grpcPort.AdServiceClient, context.Context) {
	lis := bufconn.Listen(1024 * 1024)
	t.Cleanup(func() {
		lis.Close()
	})

	srv := grpc.NewServer(grpc.ChainUnaryInterceptor(grpcPort.UnaryInterceptor, grpcPort.RecoveryInterceptor))
	t.Cleanup(func() {
		srv.Stop()
	})

	svc := grpcPort.NewService(app.NewApp(adrepo.New(), customer.New(), adfilter.New()))
	grpcPort.RegisterAdServiceServer(srv, svc)

	go func() {
		assert.NoError(t, srv.Serve(lis), "srv.Serve")
	}()

	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(func() {
		cancel()
	})

	conn, err := grpc.DialContext(ctx, "", grpc.WithContextDialer(dialer), grpc.WithInsecure())
	assert.NoError(t, err, "grpc.DialContext")

	t.Cleanup(func() {
		conn.Close()
	})

	client := grpcPort.NewAdServiceClient(conn)
	return client, ctx
}

func TestGRPCCreateUser(t *testing.T) {
	client, ctx := GetClient(t)

	res, err := client.CreateUser(ctx, &grpcPort.UniversalUser{Nickname: "Tom", Email: "example@mail.com", UserId: 3})
	assert.NoError(t, err, "client.CreateUser")
	assert.Equal(t, "Tom", res.Nickname)
	assert.Equal(t, "example@mail.com", res.Email)
	assert.Equal(t, int64(3), res.UserId)

	res, err = client.CreateUser(ctx, &grpcPort.UniversalUser{Nickname: "abc", Email: "cat@mail.com", UserId: 3})
	assert.ErrorIs(t, err, ErrorBadRequest)

	res, err = client.CreateUser(ctx, &grpcPort.UniversalUser{Nickname: "qwerty", Email: "qwerty@mail.com", UserId: 5})
	assert.NoError(t, err, "client.CreateUser")
	assert.Equal(t, "qwerty", res.Nickname)
	assert.Equal(t, "qwerty@mail.com", res.Email)
	assert.Equal(t, int64(5), res.UserId)
}

func TestGRPCCreateAd(t *testing.T) {
	client, ctx := GetClient(t)

	a, err := client.CreateUser(ctx, &grpcPort.UniversalUser{Nickname: "Tom", Email: "example@mail.com", UserId: 3})
	assert.NoError(t, err, "client.CreateUser")

	res, err := client.CreateAd(ctx, &grpcPort.CreateAdRequest{Title: "cat", Text: "text", UserId: a.UserId})
	assert.NoError(t, err, "client.CreateAd")
	assert.Equal(t, "cat", res.Title)
	assert.Equal(t, "text", res.Text)
	assert.Equal(t, a.UserId, res.AuthorId)
	assert.Equal(t, false, res.Published)

	_, err = client.CreateAd(ctx, &grpcPort.CreateAdRequest{Title: "cat", Text: "text", UserId: 5})
	assert.ErrorIs(t, err, ErrorBadRequest)
}

func TestGRPCChangeAdStatus(t *testing.T) {
	client, ctx := GetClient(t)

	a, _ := client.CreateUser(ctx, &grpcPort.UniversalUser{Nickname: "Tom", Email: "example@mail.com", UserId: 3})
	b, _ := client.CreateUser(ctx, &grpcPort.UniversalUser{Nickname: "cat", Email: "cat@mail.com", UserId: 5})

	ad, err := client.CreateAd(ctx, &grpcPort.CreateAdRequest{Title: "aba", Text: "caba", UserId: a.UserId})
	_, err = client.ChangeAdStatus(ctx, &grpcPort.ChangeAdStatusRequest{
		AdId: ad.Id, UserId: b.UserId, Published: ad.Published})
	assert.ErrorIs(t, err, ErrorForbidden)
	_, err = client.ChangeAdStatus(ctx, &grpcPort.ChangeAdStatusRequest{
		AdId: ad.Id + 1, UserId: ad.AuthorId, Published: ad.Published})
	assert.ErrorIs(t, err, ErrorBadRequest)

	updatedAd, err := client.ChangeAdStatus(ctx, &grpcPort.ChangeAdStatusRequest{
		AdId: ad.Id, UserId: ad.AuthorId, Published: true})
	assert.NoError(t, err, "client.ChangeAdStatus")
	assert.Equal(t, ad.Title, updatedAd.Title)
	assert.Equal(t, ad.Text, updatedAd.Text)
	assert.Equal(t, ad.AuthorId, updatedAd.AuthorId)
	assert.Equal(t, true, updatedAd.Published)
}

func TestGRPCUpdateAd(t *testing.T) {
	client, ctx := GetClient(t)

	a, _ := client.CreateUser(ctx, &grpcPort.UniversalUser{Nickname: "Tom", Email: "example@mail.com", UserId: 3})
	b, _ := client.CreateUser(ctx, &grpcPort.UniversalUser{Nickname: "cat", Email: "cat@mail.com", UserId: 5})

	ad, err := client.CreateAd(ctx, &grpcPort.CreateAdRequest{Title: "aba", Text: "caba", UserId: a.UserId})
	_, err = client.UpdateAd(ctx, &grpcPort.UpdateAdRequest{
		AdId: ad.Id, Title: "new title", Text: "new text", UserId: b.UserId})
	assert.ErrorIs(t, err, ErrorForbidden)
	_, err = client.UpdateAd(ctx, &grpcPort.UpdateAdRequest{
		AdId: ad.Id + 1, Title: "new title", Text: "new text", UserId: ad.AuthorId})
	assert.ErrorIs(t, err, ErrorBadRequest)

	updatedAd, err := client.UpdateAd(ctx, &grpcPort.UpdateAdRequest{
		AdId: ad.Id, Title: "new title", Text: "new text", UserId: ad.AuthorId})
	assert.NoError(t, err, "client.UpdateAd")
	assert.Equal(t, "new title", updatedAd.Title)
	assert.Equal(t, "new text", updatedAd.Text)
	assert.Equal(t, ad.AuthorId, updatedAd.AuthorId)
	assert.Equal(t, ad.Published, updatedAd.Published)
}

func TestGRPCDeleteAd(t *testing.T) {
	client, ctx := GetClient(t)

	a, _ := client.CreateUser(ctx, &grpcPort.UniversalUser{Nickname: "Tom", Email: "example@mail.com", UserId: 3})
	b, _ := client.CreateUser(ctx, &grpcPort.UniversalUser{Nickname: "cat", Email: "cat@mail.com", UserId: 5})

	ad, err := client.CreateAd(ctx, &grpcPort.CreateAdRequest{Title: "aba", Text: "caba", UserId: a.UserId})
	_, err = client.DeleteAd(ctx, &grpcPort.DeleteAdRequest{
		AdId: ad.Id, UserId: b.UserId})
	assert.ErrorIs(t, err, ErrorForbidden)
	_, err = client.DeleteAd(ctx, &grpcPort.DeleteAdRequest{
		AdId: ad.Id + 1, UserId: ad.AuthorId})
	assert.ErrorIs(t, err, ErrorBadRequest)

	resp, err := client.DeleteAd(ctx, &grpcPort.DeleteAdRequest{
		AdId: ad.Id, UserId: ad.AuthorId})
	assert.NoError(t, err, "client.DeleteAd")
	assert.Equal(t, ad.Title, resp.Title)
	assert.Equal(t, ad.Text, resp.Text)
	assert.Equal(t, ad.AuthorId, resp.AuthorId)
	assert.Equal(t, ad.Published, resp.Published)
}

func TestGRPCGetAdByID(t *testing.T) {
	client, ctx := GetClient(t)

	a, _ := client.CreateUser(ctx, &grpcPort.UniversalUser{Nickname: "Tom", Email: "example@mail.com", UserId: 3})

	ad, err := client.CreateAd(ctx, &grpcPort.CreateAdRequest{Title: "aba", Text: "caba", UserId: a.UserId})
	_, err = client.GetAdByID(ctx, &grpcPort.GetAdRequest{
		Id: ad.Id + 1})
	assert.ErrorIs(t, err, ErrorBadRequest)

	resp, err := client.GetAdByID(ctx, &grpcPort.GetAdRequest{
		Id: ad.Id})
	assert.NoError(t, err, "client.GetAdByID")
	assert.Equal(t, ad.Title, resp.Title)
	assert.Equal(t, ad.Text, resp.Text)
	assert.Equal(t, ad.AuthorId, resp.AuthorId)
	assert.Equal(t, ad.Published, resp.Published)
}

func TestGRPCDeleteUserByID(t *testing.T) {
	client, ctx := GetClient(t)

	a, _ := client.CreateUser(ctx, &grpcPort.UniversalUser{Nickname: "Tom", Email: "example@mail.com", UserId: 3})

	ad, err := client.CreateAd(ctx, &grpcPort.CreateAdRequest{Title: "aba", Text: "caba", UserId: a.UserId})
	_, err = client.DeleteUserByID(ctx, &grpcPort.DeleteUserRequest{
		Id: a.UserId + 1})
	assert.ErrorIs(t, err, ErrorBadRequest)
	_, err = client.GetAdByID(ctx, &grpcPort.GetAdRequest{
		Id: ad.Id})
	assert.NoError(t, err, "client.GetAdByID")

	resp, err := client.DeleteUserByID(ctx, &grpcPort.DeleteUserRequest{
		Id: a.UserId})
	assert.NoError(t, err, "client.DeleteAd")
	assert.Equal(t, a.Nickname, resp.Nickname)
	assert.Equal(t, a.Email, resp.Email)
	assert.Equal(t, a.UserId, resp.UserId)

	_, err = client.GetAdByID(ctx, &grpcPort.GetAdRequest{
		Id: ad.Id})
	assert.ErrorIs(t, err, ErrorBadRequest)
}

func TestGRPCChangeUserInfo(t *testing.T) {
	client, ctx := GetClient(t)

	a, _ := client.CreateUser(ctx, &grpcPort.UniversalUser{Nickname: "Tom", Email: "example@mail.com", UserId: 3})

	_, err := client.ChangeUserInfo(ctx, &grpcPort.UniversalUser{
		UserId:   a.UserId + 1,
		Nickname: "Qwerty",
		Email:    "qwerty@mail.ru",
	})
	assert.ErrorIs(t, err, ErrorBadRequest)

	resp, err := client.ChangeUserInfo(ctx, &grpcPort.UniversalUser{
		UserId:   a.UserId,
		Nickname: "Qwerty",
		Email:    "qwerty@mail.ru",
	})
	assert.NoError(t, err, "client.ChangeUserInfo")
	assert.Equal(t, "Qwerty", resp.Nickname)
	assert.Equal(t, "qwerty@mail.ru", resp.Email)
	assert.Equal(t, a.UserId, resp.UserId)
}

func TestGRPCGetAdsByTitle(t *testing.T) {
	client, ctx := GetClient(t)

	_, _ = client.CreateUser(ctx, &grpcPort.UniversalUser{Nickname: "Tom", Email: "example@mail.com", UserId: 3})
	a, _ := client.CreateAd(ctx, &grpcPort.CreateAdRequest{Title: "aba", Text: "caba", UserId: 3})
	b, _ := client.CreateAd(ctx, &grpcPort.CreateAdRequest{Title: "abacaba", Text: "12345", UserId: 3})
	_, _ = client.CreateAd(ctx, &grpcPort.CreateAdRequest{Title: "cat", Text: "text", UserId: 3})

	resp, err := client.GetAdsByTitle(ctx, &grpcPort.AdsByTitleRequest{Title: "aba"})
	assert.NoError(t, err, "client.GetAdsByTitle")
	assert.Equal(t, a.Title, resp.List[0].Title)
	assert.Equal(t, a.Text, resp.List[0].Text)
	assert.Equal(t, a.AuthorId, resp.List[0].AuthorId)
	assert.Equal(t, a.Published, resp.List[0].Published)
	assert.Equal(t, b.Title, resp.List[1].Title)
	assert.Equal(t, b.Text, resp.List[1].Text)
	assert.Equal(t, b.AuthorId, resp.List[1].AuthorId)
	assert.Equal(t, b.Published, resp.List[1].Published)
}

func TestGRPCGetUserByID(t *testing.T) {
	client, ctx := GetClient(t)

	a, _ := client.CreateUser(ctx, &grpcPort.UniversalUser{Nickname: "Tom", Email: "example@mail.com", UserId: 3})

	_, err := client.GetUserByID(ctx, &grpcPort.GetUserRequest{
		Id: a.UserId + 1,
	})
	assert.ErrorIs(t, err, ErrorNotFound)

	resp, err := client.GetUserByID(ctx, &grpcPort.GetUserRequest{
		Id: a.UserId,
	})
	assert.NoError(t, err, "client.GetUserByID")
	assert.Equal(t, a.Nickname, resp.Nickname)
	assert.Equal(t, a.Email, resp.Email)
	assert.Equal(t, a.UserId, resp.UserId)
}

func TestGRPCDefaultFilter(t *testing.T) {
	client, ctx := GetClient(t)

	_, _ = client.CreateUser(ctx, &grpcPort.UniversalUser{UserId: 123,
		Nickname: "nickname", Email: "example@mail.com"})

	resp, err := client.CreateAd(ctx, &grpcPort.CreateAdRequest{UserId: 123,
		Title: "hello", Text: "world"})
	assert.NoError(t, err, "client.CreateAd")

	publishedAd, err := client.ChangeAdStatus(ctx, &grpcPort.ChangeAdStatusRequest{
		UserId: 123, AdId: resp.Id, Published: true})
	assert.NoError(t, err, "client.ChangeAdStatus")

	_, err = client.CreateAd(ctx, &grpcPort.CreateAdRequest{UserId: 123,
		Title: "best cat", Text: "not for sale"})
	assert.NoError(t, err, "client.CreateAd")

	ads, err := client.ListAds(ctx, &grpcPort.FilterRequest{})
	assert.NoError(t, err, "client.ListAds")
	assert.Len(t, ads.List, 1)
	assert.Equal(t, ads.List[0].Id, publishedAd.Id)
	assert.Equal(t, ads.List[0].Title, publishedAd.Title)
	assert.Equal(t, ads.List[0].Text, publishedAd.Text)
	assert.Equal(t, ads.List[0].AuthorId, publishedAd.AuthorId)
	assert.True(t, ads.List[0].Published)
}

func TestGRPCFilterByAuthor(t *testing.T) {
	client, ctx := GetClient(t)

	_, _ = client.CreateUser(ctx, &grpcPort.UniversalUser{UserId: 3,
		Nickname: "nickname", Email: "example@mail.com"})
	_, _ = client.CreateUser(ctx, &grpcPort.UniversalUser{UserId: 5,
		Nickname: "cat", Email: "cat@mail.com"})
	_, _ = client.CreateUser(ctx, &grpcPort.UniversalUser{UserId: 7,
		Nickname: "aba", Email: "caba@mail.com"})

	a, _ := client.CreateAd(ctx, &grpcPort.CreateAdRequest{UserId: 3,
		Title: "aba", Text: "caba"})
	b, _ := client.CreateAd(ctx, &grpcPort.CreateAdRequest{UserId: 3,
		Title: "bab", Text: "abac"})
	c, _ := client.CreateAd(ctx, &grpcPort.CreateAdRequest{UserId: 5,
		Title: "foo", Text: "bar"})
	d, _ := client.CreateAd(ctx, &grpcPort.CreateAdRequest{UserId: 7,
		Title: "alpha", Text: "beta"})

	a, _ = client.ChangeAdStatus(ctx, &grpcPort.ChangeAdStatusRequest{
		UserId: 3, AdId: a.Id, Published: true})
	b, _ = client.ChangeAdStatus(ctx, &grpcPort.ChangeAdStatusRequest{
		UserId: 3, AdId: b.Id, Published: true})
	c, _ = client.ChangeAdStatus(ctx, &grpcPort.ChangeAdStatusRequest{
		UserId: 5, AdId: c.Id, Published: true})
	d, _ = client.ChangeAdStatus(ctx, &grpcPort.ChangeAdStatusRequest{
		UserId: 7, AdId: d.Id, Published: true})

	ads, err := client.ListAds(ctx, &grpcPort.FilterRequest{AuthorId: 3})
	assert.NoError(t, err)
	assert.Equal(t, ads.List[0].Id, a.Id)
	assert.Equal(t, ads.List[0].Title, a.Title)
	assert.Equal(t, ads.List[0].Text, a.Text)
	assert.Equal(t, ads.List[0].AuthorId, a.AuthorId)
	assert.Equal(t, ads.List[0].Published, a.Published)

	assert.Equal(t, ads.List[1].Id, b.Id)
	assert.Equal(t, ads.List[1].Title, b.Title)
	assert.Equal(t, ads.List[1].Text, b.Text)
	assert.Equal(t, ads.List[1].AuthorId, b.AuthorId)
	assert.Equal(t, ads.List[1].Published, b.Published)
}

func TestGRPCFilterByTime(t *testing.T) {
	client, ctx := GetClient(t)

	_, _ = client.CreateUser(ctx, &grpcPort.UniversalUser{UserId: 3,
		Nickname: "nickname", Email: "example@mail.com"})
	_, _ = client.CreateUser(ctx, &grpcPort.UniversalUser{UserId: 5,
		Nickname: "cat", Email: "cat@mail.com"})
	_, _ = client.CreateUser(ctx, &grpcPort.UniversalUser{UserId: 7,
		Nickname: "aba", Email: "caba@mail.com"})

	a, _ := client.CreateAd(ctx, &grpcPort.CreateAdRequest{UserId: 3,
		Title: "aba", Text: "caba"})
	time.Sleep(2 * time.Second)
	tm := time.Now()
	lTm := timestamppb.New(tm.Add(-time.Second))
	rTm := timestamppb.New(tm.Add(time.Second))
	b, _ := client.CreateAd(ctx, &grpcPort.CreateAdRequest{UserId: 3,
		Title: "bab", Text: "abac"})
	c, _ := client.CreateAd(ctx, &grpcPort.CreateAdRequest{UserId: 5,
		Title: "foo", Text: "bar"})
	time.Sleep(2 * time.Second)
	d, _ := client.CreateAd(ctx, &grpcPort.CreateAdRequest{UserId: 7,
		Title: "alpha", Text: "beta"})

	a, _ = client.ChangeAdStatus(ctx, &grpcPort.ChangeAdStatusRequest{
		UserId: 3, AdId: a.Id, Published: true})
	b, _ = client.ChangeAdStatus(ctx, &grpcPort.ChangeAdStatusRequest{
		UserId: 3, AdId: b.Id, Published: true})
	c, _ = client.ChangeAdStatus(ctx, &grpcPort.ChangeAdStatusRequest{
		UserId: 5, AdId: c.Id, Published: true})
	d, _ = client.ChangeAdStatus(ctx, &grpcPort.ChangeAdStatusRequest{
		UserId: 7, AdId: d.Id, Published: true})

	ads, err := client.ListAds(ctx, &grpcPort.FilterRequest{LDate: lTm, RDate: rTm})
	assert.NoError(t, err)
	assert.Equal(t, ads.List[0].Id, b.Id)
	assert.Equal(t, ads.List[0].Title, b.Title)
	assert.Equal(t, ads.List[0].Text, b.Text)
	assert.Equal(t, ads.List[0].AuthorId, b.AuthorId)
	assert.Equal(t, ads.List[0].Published, b.Published)

	assert.Equal(t, ads.List[1].Id, c.Id)
	assert.Equal(t, ads.List[1].Title, c.Title)
	assert.Equal(t, ads.List[1].Text, c.Text)
	assert.Equal(t, ads.List[1].AuthorId, c.AuthorId)
	assert.Equal(t, ads.List[1].Published, c.Published)
}
