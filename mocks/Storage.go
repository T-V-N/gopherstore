// Code generated by mockery v2.15.0. DO NOT EDIT.

package mocks

import (
	context "context"

	sharedTypes "github.com/T-V-N/gopherstore/internal/shared_types"
	mock "github.com/stretchr/testify/mock"
)

// Storage is an autogenerated mock type for the Storage type
type Storage struct {
	mock.Mock
}

// CreateOrder provides a mock function with given fields: ctx, orderID, uid
func (_m *Storage) CreateOrder(ctx context.Context, orderID string, uid string) error {
	ret := _m.Called(ctx, orderID, uid)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, orderID, uid)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateUser provides a mock function with given fields: ctx, creds
func (_m *Storage) CreateUser(ctx context.Context, creds sharedTypes.Credentials) (string, error) {
	ret := _m.Called(ctx, creds)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, sharedTypes.Credentials) string); ok {
		r0 = rf(ctx, creds)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, sharedTypes.Credentials) error); ok {
		r1 = rf(ctx, creds)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetBalance provides a mock function with given fields: ctx, uid
func (_m *Storage) GetBalance(ctx context.Context, uid string) (sharedTypes.Balance, error) {
	ret := _m.Called(ctx, uid)

	var r0 sharedTypes.Balance
	if rf, ok := ret.Get(0).(func(context.Context, string) sharedTypes.Balance); ok {
		r0 = rf(ctx, uid)
	} else {
		r0 = ret.Get(0).(sharedTypes.Balance)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, uid)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetListWithdrawals provides a mock function with given fields: ctx, uid
func (_m *Storage) GetListWithdrawals(ctx context.Context, uid string) ([]sharedTypes.Withdrawal, error) {
	ret := _m.Called(ctx, uid)

	var r0 []sharedTypes.Withdrawal
	if rf, ok := ret.Get(0).(func(context.Context, string) []sharedTypes.Withdrawal); ok {
		r0 = rf(ctx, uid)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]sharedTypes.Withdrawal)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, uid)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetUser provides a mock function with given fields: ctx, creds
func (_m *Storage) GetUser(ctx context.Context, creds sharedTypes.Credentials) (string, error) {
	ret := _m.Called(ctx, creds)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, sharedTypes.Credentials) string); ok {
		r0 = rf(ctx, creds)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, sharedTypes.Credentials) error); ok {
		r1 = rf(ctx, creds)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListOrders provides a mock function with given fields: ctx, uid
func (_m *Storage) ListOrders(ctx context.Context, uid string) ([]sharedTypes.Order, error) {
	ret := _m.Called(ctx, uid)

	var r0 []sharedTypes.Order
	if rf, ok := ret.Get(0).(func(context.Context, string) []sharedTypes.Order); ok {
		r0 = rf(ctx, uid)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]sharedTypes.Order)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, uid)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// WithdrawBalance provides a mock function with given fields: ctx, uid, orderID, amount
func (_m *Storage) WithdrawBalance(ctx context.Context, uid string, orderID string, amount float32) error {
	ret := _m.Called(ctx, uid, orderID, amount)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, float32) error); ok {
		r0 = rf(ctx, uid, orderID, amount)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewStorage interface {
	mock.TestingT
	Cleanup(func())
}

// NewStorage creates a new instance of Storage. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewStorage(t mockConstructorTestingTNewStorage) *Storage {
	mock := &Storage{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
