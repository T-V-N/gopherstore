// Code generated by mockery v2.15.0. DO NOT EDIT.

package mocks

import (
	context "context"

	sharedtypes "github.com/T-V-N/gopherstore/internal/shared_types"
	mock "github.com/stretchr/testify/mock"
)

// WithdrawalApper is an autogenerated mock type for the WithdrawalApper type
type WithdrawalApper struct {
	mock.Mock
}

// GetListWithdrawals provides a mock function with given fields: ctx, uid
func (_m *WithdrawalApper) GetListWithdrawals(ctx context.Context, uid string) ([]sharedtypes.Withdrawal, error) {
	ret := _m.Called(ctx, uid)

	var r0 []sharedtypes.Withdrawal
	if rf, ok := ret.Get(0).(func(context.Context, string) []sharedtypes.Withdrawal); ok {
		r0 = rf(ctx, uid)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]sharedtypes.Withdrawal)
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

type mockConstructorTestingTNewWithdrawalApper interface {
	mock.TestingT
	Cleanup(func())
}

// NewWithdrawalApper creates a new instance of WithdrawalApper. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewWithdrawalApper(t mockConstructorTestingTNewWithdrawalApper) *WithdrawalApper {
	mock := &WithdrawalApper{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
