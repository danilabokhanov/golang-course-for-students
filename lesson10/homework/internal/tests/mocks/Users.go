// Code generated by mockery v2.26.1. DO NOT EDIT.

package mocks

import (
	context "context"
	user "homework10/internal/user"

	mock "github.com/stretchr/testify/mock"
)

// Users is an autogenerated mock type for the Users type
type Users struct {
	mock.Mock
}

// ChangeInfo provides a mock function with given fields: ctx, userID, nickname, email
func (_m *Users) ChangeInfo(ctx context.Context, userID int64, nickname string, email string) error {
	ret := _m.Called(ctx, userID, nickname, email)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int64, string, string) error); ok {
		r0 = rf(ctx, userID, nickname, email)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateByID provides a mock function with given fields: ctx, nickname, email, userID
func (_m *Users) CreateByID(ctx context.Context, nickname string, email string, userID int64) (user.User, error) {
	ret := _m.Called(ctx, nickname, email, userID)

	var r0 user.User
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, int64) (user.User, error)); ok {
		return rf(ctx, nickname, email, userID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, int64) user.User); ok {
		r0 = rf(ctx, nickname, email, userID)
	} else {
		r0 = ret.Get(0).(user.User)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, int64) error); ok {
		r1 = rf(ctx, nickname, email, userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeleteByID provides a mock function with given fields: ctx, userID
func (_m *Users) DeleteByID(ctx context.Context, userID int64) (user.User, error) {
	ret := _m.Called(ctx, userID)

	var r0 user.User
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64) (user.User, error)); ok {
		return rf(ctx, userID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64) user.User); ok {
		r0 = rf(ctx, userID)
	} else {
		r0 = ret.Get(0).(user.User)
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64) error); ok {
		r1 = rf(ctx, userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Find provides a mock function with given fields: ctx, userID
func (_m *Users) Find(ctx context.Context, userID int64) (user.User, bool) {
	ret := _m.Called(ctx, userID)

	var r0 user.User
	var r1 bool
	if rf, ok := ret.Get(0).(func(context.Context, int64) (user.User, bool)); ok {
		return rf(ctx, userID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64) user.User); ok {
		r0 = rf(ctx, userID)
	} else {
		r0 = ret.Get(0).(user.User)
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64) bool); ok {
		r1 = rf(ctx, userID)
	} else {
		r1 = ret.Get(1).(bool)
	}

	return r0, r1
}

type mockConstructorTestingTNewUsers interface {
	mock.TestingT
	Cleanup(func())
}

// NewUsers creates a new instance of Users. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewUsers(t mockConstructorTestingTNewUsers) *Users {
	mock := &Users{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
