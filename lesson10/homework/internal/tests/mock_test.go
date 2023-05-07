package tests

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"homework10/internal/adapters/adfilter"
	"homework10/internal/adapters/adrepo"
	"homework10/internal/adapters/customer"
	"homework10/internal/adpattern"
	"homework10/internal/ads"
	"homework10/internal/app"
	grpcPort "homework10/internal/ports/grpc"
	"homework10/internal/tests/mocks"
	"homework10/internal/user"
	"net"
	"testing"
	"time"
)

func getGRPCClient(t *testing.T, a app.App) (grpcPort.AdServiceClient, context.Context) {
	lis := bufconn.Listen(1024 * 1024)
	t.Cleanup(func() {
		lis.Close()
	})

	srv := grpc.NewServer(grpc.ChainUnaryInterceptor(grpcPort.UnaryInterceptor, grpcPort.RecoveryInterceptor))
	t.Cleanup(func() {
		srv.Stop()
	})

	svc := grpcPort.NewService(a)
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

	conn, err := grpc.DialContext(ctx, "", grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err, "grpc.DialContext")

	t.Cleanup(func() {
		conn.Close()
	})

	client := grpcPort.NewAdServiceClient(conn)
	return client, ctx
}

func Test_CreateAd(t *testing.T) {
	repo := &mocks.Repository{}
	repo.On("Add", mock.AnythingOfType("*context.emptyCtx"),
		mock.AnythingOfType("string"), mock.AnythingOfType("string"),
		mock.AnythingOfType("int64")).
		Return(int64(0), fmt.Errorf("add error")).Once()

	a := app.NewApp(repo, customer.New(), adfilter.New())
	ctx := context.Background()
	_, _ = a.CreateUserByID(ctx, "first user", "example@mail.ru", 1)
	_, err := a.CreateAd(ctx, "test ad", "abacaba", int64(1))
	assert.ErrorIs(t, err, app.ErrApp)
}

func Test_DeleteAd(t *testing.T) {
	repo := &mocks.Repository{}
	userId := int64(1)
	repo.On("Find", mock.AnythingOfType("*context.emptyCtx"),
		mock.AnythingOfType("int64")).
		Return(ads.Ad{AuthorID: userId}, true).Once()
	repo.On("Delete", mock.AnythingOfType("*context.emptyCtx"),
		mock.AnythingOfType("int64")).
		Return(fmt.Errorf("delete error")).Once()

	a := app.NewApp(repo, customer.New(), adfilter.New())
	ctx := context.Background()
	_, err := a.DeleteAd(ctx, 1, userId)
	assert.ErrorIs(t, err, app.ErrApp)
}

func Test_ChangeAdStatus(t *testing.T) {
	repo := &mocks.Repository{}
	userId := int64(1)
	repo.On("Find", mock.AnythingOfType("*context.emptyCtx"),
		mock.AnythingOfType("int64")).
		Return(ads.Ad{AuthorID: userId}, true).Once()
	repo.On("SetStatus", mock.AnythingOfType("*context.emptyCtx"),
		mock.AnythingOfType("int64"), mock.AnythingOfType("bool")).
		Return(fmt.Errorf("set status error")).Once()

	a := app.NewApp(repo, customer.New(), adfilter.New())
	ctx := context.Background()
	_, _ = a.CreateUserByID(ctx, "test user", "example@mail.ru", userId)
	_, err := a.ChangeAdStatus(ctx, 1, userId, true)
	assert.ErrorIs(t, err, app.ErrApp)

	_, err = a.ChangeAdStatus(ctx, 1, userId+1, true)
	assert.ErrorIs(t, err, app.ErrWrongFormat)
}

func Test_UpdateAd(t *testing.T) {
	repo := &mocks.Repository{}
	userId := int64(1)
	repo.On("Find", mock.AnythingOfType("*context.emptyCtx"),
		mock.AnythingOfType("int64")).
		Return(ads.Ad{AuthorID: userId}, true)
	repo.On("SetText", mock.AnythingOfType("*context.emptyCtx"),
		mock.AnythingOfType("int64"), mock.AnythingOfType("string")).
		Return(fmt.Errorf("set text error")).Once()

	a := app.NewApp(repo, customer.New(), adfilter.New())
	ctx := context.Background()
	_, _ = a.CreateUserByID(ctx, "test user", "example@mail.ru", userId)
	_, err := a.UpdateAd(ctx, 1, userId, "aba", "caba")
	assert.ErrorIs(t, err, app.ErrApp)

	_, err = a.UpdateAd(ctx, 1, userId+1, "aba", "caba")
	assert.ErrorIs(t, err, app.ErrWrongFormat)

	repo.On("SetText", mock.AnythingOfType("*context.emptyCtx"),
		mock.AnythingOfType("int64"), mock.AnythingOfType("string")).
		Return(nil).Once()
	repo.On("SetTitle", mock.AnythingOfType("*context.emptyCtx"),
		mock.AnythingOfType("int64"), mock.AnythingOfType("string")).
		Return(fmt.Errorf("set title error")).Once()
	_, err = a.UpdateAd(ctx, 1, userId, "aba", "caba")
	assert.ErrorIs(t, err, app.ErrApp)
}

func Test_ChangeUserInfo(t *testing.T) {
	u := &mocks.Users{}
	userId := int64(1)
	u.On("Find", mock.AnythingOfType("*context.emptyCtx"),
		mock.AnythingOfType("int64")).
		Return(user.User{}, true).Once()
	u.On("ChangeInfo", mock.AnythingOfType("*context.emptyCtx"),
		mock.AnythingOfType("int64"), mock.AnythingOfType("string"),
		mock.AnythingOfType("string")).
		Return(fmt.Errorf("change info error")).Once()

	a := app.NewApp(adrepo.New(), u, adfilter.New())
	ctx := context.Background()
	_, err := a.ChangeUserInfo(ctx, userId, "nickname", "example@mail.ru")
	assert.ErrorIs(t, err, app.ErrApp)
}

func Test_GetAllAdsByTemplate(t *testing.T) {
	repo := &mocks.Repository{}
	repo.On("GetAllByTemplate", mock.AnythingOfType("*context.emptyCtx"),
		mock.AnythingOfType("adpattern.AdPattern")).
		Return([]ads.Ad{}, fmt.Errorf("get all by template error")).Once()

	a := app.NewApp(repo, customer.New(), adfilter.New())
	ctx := context.Background()
	_, err := a.GetAllAdsByTemplate(ctx, adpattern.AdPattern{})
	assert.ErrorIs(t, err, app.ErrApp)
}

func Test_GetNewFilter(t *testing.T) {
	f := &mocks.Filter{}
	f.On("BasicConfig", mock.AnythingOfType("*context.emptyCtx")).
		Return(f, fmt.Errorf("basic config error")).Once()

	a := app.NewApp(adrepo.New(), customer.New(), f)
	ctx := context.Background()
	_, err := a.GetNewFilter(ctx)
	assert.ErrorIs(t, err, app.ErrApp)
}

func Test_GetAdsByTitle(t *testing.T) {
	repo := &mocks.Repository{}
	repo.On("GetByTitle", mock.AnythingOfType("*context.emptyCtx"),
		mock.AnythingOfType("string")).
		Return([]ads.Ad{}, fmt.Errorf("get by title error")).Once()

	a := app.NewApp(repo, customer.New(), adfilter.New())
	ctx := context.Background()
	_, err := a.GetAdsByTitle(ctx, "abacaba")
	assert.ErrorIs(t, err, app.ErrApp)
}

func Test_CreateUserByID(t *testing.T) {
	u := &mocks.Users{}
	userId := int64(1)
	u.On("Find", mock.AnythingOfType("*context.emptyCtx"),
		mock.AnythingOfType("int64")).
		Return(user.User{}, false).Once()
	u.On("CreateByID", mock.AnythingOfType("*context.emptyCtx"),
		mock.AnythingOfType("string"), mock.AnythingOfType("string"),
		mock.AnythingOfType("int64")).
		Return(user.User{}, fmt.Errorf("create by id error")).Once()

	a := app.NewApp(adrepo.New(), u, adfilter.New())
	ctx := context.Background()
	_, err := a.CreateUserByID(ctx, "nickname", "example@mail.ru", userId)
	assert.ErrorIs(t, err, app.ErrApp)
}

func Test_DeleteUserByID(t *testing.T) {
	u := &mocks.Users{}
	userId := int64(1)
	u.On("Find", mock.AnythingOfType("*context.emptyCtx"),
		mock.AnythingOfType("int64")).
		Return(user.User{}, true)
	u.On("DeleteByID", mock.AnythingOfType("*context.emptyCtx"),
		mock.AnythingOfType("int64")).
		Return(user.User{}, fmt.Errorf("delete by id error")).Once()

	repo := &mocks.Repository{}
	a := app.NewApp(repo, u, adfilter.New())
	ctx := context.Background()
	_, err := a.DeleteUserByID(ctx, userId)
	assert.ErrorIs(t, err, app.ErrApp)

	u.On("DeleteByID", mock.AnythingOfType("*context.emptyCtx"),
		mock.AnythingOfType("int64")).
		Return(user.User{}, nil).Once()

	repo.On("DeleteByAuthor", mock.AnythingOfType("*context.emptyCtx"),
		mock.AnythingOfType("int64")).
		Return(fmt.Errorf("delete by author error")).Once()
	_, err = a.DeleteUserByID(ctx, userId)
	assert.ErrorIs(t, err, app.ErrApp)
}

func Test_BrokenApp(t *testing.T) {
	testApp := mocks.App{}
	testApp.On("CreateAd", mock.AnythingOfType("*gin.Context"),
		mock.AnythingOfType("string"), mock.AnythingOfType("string"),
		mock.AnythingOfType("int64")).
		Return(ads.Ad{}, app.ErrApp)
	testApp.On("GetNewFilter", mock.AnythingOfType("*gin.Context")).
		Return(&adfilter.BasicFilter{}, app.ErrApp)
	testApp.On("GetAdsByTitle", mock.AnythingOfType("*gin.Context"),
		mock.AnythingOfType("string")).
		Return([]ads.Ad{}, app.ErrApp)
	client := getTestClient(&testApp)
	_, err := client.createAd(3, "aba", "caba")
	assert.ErrorIs(t, err, InternalServerErr)
	_, err = client.listAdsBasic()
	assert.ErrorIs(t, err, InternalServerErr)
	_, err = client.getAdsByTitle("aba")
	assert.ErrorIs(t, err, InternalServerErr)
}

func Test_changeUserInfo(t *testing.T) {
	testApp := mocks.App{}
	testApp.On("FindUser", mock.AnythingOfType("*gin.Context"),
		mock.AnythingOfType("int64")).
		Return(user.User{}, false, app.ErrApp).Once()
	client := getTestClient(&testApp)

	_, err := client.changeUserInfo(3, "nickname", "example@mail.ru")
	assert.ErrorIs(t, err, InternalServerErr)

	testApp.On("FindUser", mock.AnythingOfType("*gin.Context"),
		mock.AnythingOfType("int64")).
		Return(user.User{}, true, nil)
	testApp.On("ChangeUserInfo", mock.AnythingOfType("*gin.Context"),
		mock.AnythingOfType("int64"), mock.AnythingOfType("string"),
		mock.AnythingOfType("string")).
		Return(user.User{}, app.ErrWrongFormat).Once()

	_, err = client.changeUserInfo(3, "nickname", "example@mail.ru")
	assert.ErrorIs(t, err, ErrBadRequest)

	testApp.On("ChangeUserInfo", mock.AnythingOfType("*gin.Context"),
		mock.AnythingOfType("int64"), mock.AnythingOfType("string"),
		mock.AnythingOfType("string")).
		Return(user.User{}, app.ErrNoAccess).Once()

	_, err = client.changeUserInfo(3, "nickname", "example@mail.ru")
	assert.ErrorIs(t, err, ErrForbidden)

	testApp.On("ChangeUserInfo", mock.AnythingOfType("*gin.Context"),
		mock.AnythingOfType("int64"), mock.AnythingOfType("string"),
		mock.AnythingOfType("string")).
		Return(user.User{}, app.ErrApp).Once()

	_, err = client.changeUserInfo(3, "nickname", "example@mail.ru")
	assert.ErrorIs(t, err, InternalServerErr)
}

func Test_listAds(t *testing.T) {
	testApp := mocks.App{}
	f := mocks.Filter{}
	f.On("BasicConfig", mock.AnythingOfType("*gin.Context")).
		Return(adfilter.New(), app.ErrApp).Once()
	testApp.On("GetNewFilter", mock.AnythingOfType("*gin.Context")).
		Return(&f, nil)
	testApp.On("GetAllAdsByTemplate", mock.AnythingOfType("*gin.Context"),
		mock.AnythingOfType("AdPattern")).
		Return([]ads.Ad{}, app.ErrApp)
	client := getTestClient(&testApp)

	_, err := client.listAdsBasic()
	assert.ErrorIs(t, err, InternalServerErr)

	f.On("BasicConfig", mock.AnythingOfType("*gin.Context")).
		Return(&f, nil)
	f.On("GetPattern", mock.AnythingOfType("*gin.Context")).
		Return(adpattern.AdPattern{}, nil).Once()
	_, err = client.listAdsBasic()
	assert.ErrorIs(t, err, InternalServerErr)

	f.On("GetPattern", mock.AnythingOfType("*gin.Context")).
		Return(adpattern.AdPattern{}, app.ErrApp).Once()
	_, err = client.listAdsBasic()
	assert.ErrorIs(t, err, InternalServerErr)
}

func Test_changeAdStatus(t *testing.T) {
	testApp := mocks.App{}
	testApp.On("FindAd", mock.AnythingOfType("*gin.Context"),
		mock.AnythingOfType("int64")).
		Return(ads.Ad{}, app.ErrApp).Once()
	client := getTestClient(&testApp)

	_, err := client.changeAdStatus(1, 1, true)
	assert.ErrorIs(t, err, ErrBadRequest)

	testApp.On("FindAd", mock.AnythingOfType("*gin.Context"),
		mock.AnythingOfType("int64")).
		Return(ads.Ad{}, nil)
	testApp.On("ChangeAdStatus", mock.AnythingOfType("*gin.Context"),
		mock.AnythingOfType("int64"), mock.AnythingOfType("int64"),
		mock.AnythingOfType("bool")).
		Return(ads.Ad{}, app.ErrWrongFormat).Once()

	_, err = client.changeAdStatus(1, 1, true)
	assert.ErrorIs(t, err, ErrBadRequest)

	testApp.On("ChangeAdStatus", mock.AnythingOfType("*gin.Context"),
		mock.AnythingOfType("int64"), mock.AnythingOfType("int64"),
		mock.AnythingOfType("bool")).
		Return(ads.Ad{}, app.ErrNoAccess).Once()

	_, err = client.changeAdStatus(1, 1, true)
	assert.ErrorIs(t, err, ErrForbidden)

	testApp.On("ChangeAdStatus", mock.AnythingOfType("*gin.Context"),
		mock.AnythingOfType("int64"), mock.AnythingOfType("int64"),
		mock.AnythingOfType("bool")).
		Return(ads.Ad{}, app.ErrApp).Once()

	_, err = client.changeAdStatus(1, 1, true)
	assert.ErrorIs(t, err, InternalServerErr)
}

func Test_getAdByID(t *testing.T) {
	testApp := mocks.App{}
	testApp.On("FindAd", mock.AnythingOfType("*gin.Context"),
		mock.AnythingOfType("int64")).
		Return(ads.Ad{}, app.ErrApp).Once()
	client := getTestClient(&testApp)

	_, err := client.getAdByID(1)
	assert.ErrorIs(t, err, InternalServerErr)
}

func Test_createUser(t *testing.T) {
	testApp := mocks.App{}
	testApp.On("CreateUserByID", mock.AnythingOfType("*gin.Context"),
		mock.AnythingOfType("string"), mock.AnythingOfType("string"),
		mock.AnythingOfType("int64")).
		Return(user.User{}, app.ErrApp).Once()
	client := getTestClient(&testApp)

	_, err := client.createUser(1, "aba", "caba@mail.ru")
	assert.ErrorIs(t, err, InternalServerErr)
}

func Test_deleteUserByID(t *testing.T) {
	testApp := mocks.App{}
	testApp.On("DeleteUserByID", mock.AnythingOfType("*gin.Context"),
		mock.AnythingOfType("int64")).
		Return(user.User{}, app.ErrApp).Once()
	client := getTestClient(&testApp)

	_, err := client.deleteUserByID(1)
	assert.ErrorIs(t, err, InternalServerErr)
}

func Test_getUserByID(t *testing.T) {
	testApp := mocks.App{}
	testApp.On("FindUser", mock.AnythingOfType("*gin.Context"),
		mock.AnythingOfType("int64")).
		Return(user.User{}, false, app.ErrApp).Once()
	client := getTestClient(&testApp)

	_, err := client.getUserByID(1)
	assert.ErrorIs(t, err, InternalServerErr)
}

func Test_ListAds(t *testing.T) {
	testApp := mocks.App{}

	client, ctx := getGRPCClient(t, &testApp)
	f := mocks.Filter{}
	testApp.On("GetNewFilter", mock.AnythingOfType("*context.valueCtx")).
		Return(&f, app.ErrApp).Once()
	_, err := client.ListAds(ctx, &grpcPort.FilterRequest{})
	assert.ErrorIs(t, err, ErrorInternal)

	testApp.On("GetNewFilter", mock.AnythingOfType("*context.valueCtx")).
		Return(&f, nil)
	f.On("SetAuthor", mock.AnythingOfType("*context.valueCtx"),
		mock.AnythingOfType("int64")).
		Return(&f, app.ErrApp).Once()

	_, err = client.ListAds(ctx, &grpcPort.FilterRequest{})
	assert.ErrorIs(t, err, ErrorInternal)

	f.On("SetAuthor", mock.AnythingOfType("*context.valueCtx"),
		mock.AnythingOfType("int64")).
		Return(&f, nil)
	f.On("SetStatus", mock.AnythingOfType("*context.valueCtx"),
		mock.AnythingOfType("bool")).
		Return(&f, app.ErrApp).Once()

	_, err = client.ListAds(ctx,
		&grpcPort.FilterRequest{PublishedConfig: grpcPort.PublishedConfig_PublishedOnly})
	assert.ErrorIs(t, err, ErrorInternal)

	f.On("GetPattern", mock.AnythingOfType("*context.valueCtx")).
		Return(adpattern.AdPattern{}, app.ErrApp).Once()

	_, err = client.ListAds(ctx, &grpcPort.FilterRequest{})
	assert.ErrorIs(t, err, ErrorInternal)

	f.On("GetPattern", mock.AnythingOfType("*context.valueCtx")).
		Return(adpattern.AdPattern{}, nil)
	testApp.On("GetAllAdsByTemplate", mock.AnythingOfType("*context.valueCtx"),
		mock.AnythingOfType("adpattern.AdPattern")).
		Return([]ads.Ad{}, app.ErrApp).Once()

	_, err = client.ListAds(ctx, &grpcPort.FilterRequest{})
	assert.ErrorIs(t, err, ErrorInternal)
}

func Test_DeleteUserById(t *testing.T) {
	testApp := mocks.App{}

	client, ctx := getGRPCClient(t, &testApp)
	testApp.On("DeleteUserByID", mock.AnythingOfType("*context.valueCtx"),
		mock.AnythingOfType("int64")).
		Return(user.User{}, app.ErrApp).Once()
	_, err := client.DeleteUserByID(ctx, &grpcPort.DeleteUserRequest{})
	assert.ErrorIs(t, err, ErrorInternal)
}

func Test_ChangeUserData(t *testing.T) {
	testApp := mocks.App{}

	client, ctx := getGRPCClient(t, &testApp)
	testApp.On("ChangeUserInfo", mock.AnythingOfType("*context.valueCtx"),
		mock.AnythingOfType("int64"), mock.AnythingOfType("string"),
		mock.AnythingOfType("string")).
		Return(user.User{}, app.ErrApp).Once()
	_, err := client.ChangeUserInfo(ctx, &grpcPort.UniversalUser{})
	assert.ErrorIs(t, err, ErrorInternal)
}

func Test_GetAdByTitle(t *testing.T) {
	testApp := mocks.App{}

	client, ctx := getGRPCClient(t, &testApp)
	testApp.On("GetAdsByTitle", mock.AnythingOfType("*context.valueCtx"),
		mock.AnythingOfType("string")).
		Return([]ads.Ad{}, app.ErrApp).Once()
	_, err := client.GetAdsByTitle(ctx, &grpcPort.AdsByTitleRequest{})
	assert.ErrorIs(t, err, ErrorInternal)
}

func Test_GetUserByID(t *testing.T) {
	testApp := mocks.App{}

	client, ctx := getGRPCClient(t, &testApp)
	testApp.On("FindUser", mock.AnythingOfType("*context.valueCtx"),
		mock.AnythingOfType("int64")).
		Return(user.User{}, false, app.ErrApp).Once()
	_, err := client.GetUserByID(ctx, &grpcPort.GetUserRequest{})
	assert.ErrorIs(t, err, ErrorInternal)
}

func Test_CreateUser(t *testing.T) {
	testApp := mocks.App{}

	client, ctx := getGRPCClient(t, &testApp)
	testApp.On("CreateUserByID", mock.AnythingOfType("*context.valueCtx"),
		mock.AnythingOfType("string"), mock.AnythingOfType("string"),
		mock.AnythingOfType("int64")).
		Return(user.User{}, app.ErrApp).Once()
	_, err := client.CreateUser(ctx, &grpcPort.UniversalUser{})
	assert.ErrorIs(t, err, ErrorInternal)
}

func Test_GetAdByID(t *testing.T) {
	testApp := mocks.App{}

	client, ctx := getGRPCClient(t, &testApp)
	testApp.On("FindAd", mock.AnythingOfType("*context.valueCtx"),
		mock.AnythingOfType("int64")).
		Return(ads.Ad{}, app.ErrApp).Once()
	_, err := client.GetAdByID(ctx, &grpcPort.GetAdRequest{})
	assert.ErrorIs(t, err, ErrorInternal)
}

func Test_InsertAd(t *testing.T) {
	testApp := mocks.App{}

	client, ctx := getGRPCClient(t, &testApp)
	testApp.On("CreateAd", mock.AnythingOfType("*context.valueCtx"),
		mock.AnythingOfType("string"), mock.AnythingOfType("string"),
		mock.AnythingOfType("int64")).
		Return(ads.Ad{}, app.ErrApp).Once()
	_, err := client.CreateAd(ctx, &grpcPort.CreateAdRequest{})
	assert.ErrorIs(t, err, ErrorInternal)
}
