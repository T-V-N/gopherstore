package sharedtypes

import (
	"context"
	"time"
)

type Credentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Order struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    float32   `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}

type Balance struct {
	Current   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}

type WtihdrawRequest struct {
	OrderID string  `json:"order"`
	Sum     float32 `json:"sum"`
}

type Withdrawal struct {
	ID          string    `json:"order"`
	Sum         float32   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

type User struct {
	UID            string
	Login          string
	PasswordHash   string
	CurrentBalance float32
	Withdrawn      float32
	CreatedAt      string
}

type UserStorager interface {
	CreateUser(context.Context, Credentials) (string, error)
	GetUser(context.Context, Credentials) (User, error)
	GetBalance(context.Context, string) (Balance, error)
	GetBalanceAndLock(context.Context, string) (Balance, error)
	WithdrawBalance(context.Context, string, string, float32, float32, float32, WithdrawalStorager) error
	UpdateUser(context.Context, string, string, float32) error
}

type OrderStorager interface {
	CreateOrder(context.Context, string, string) error
	ListOrders(context.Context, string) ([]Order, error)
	GetUnproccessedOrders(context.Context) ([]Order, error)
	UpdateOrder(context.Context, string, string, float32, UserStorager) error
}

type WithdrawalStorager interface {
	ListWithdrawals(context.Context, string) ([]Withdrawal, error)
	CreateWithdrawal(context.Context, string, float32, string) error
}

type OrderApper interface {
	CreateOrder(ctx context.Context, orderID string, uid string) error
	ListOrders(ctx context.Context, uid string) ([]Order, error)
}

type UserApper interface {
	Register(ctx context.Context, creds Credentials) (string, error)
	Login(ctx context.Context, creds Credentials) (string, error)
	GetBalance(ctx context.Context, uid string) (Balance, error)
	WithdrawBalance(ctx context.Context, uid string, orderID string, amount float32) error
}

type WithdrawalApper interface {
	GetListWithdrawals(ctx context.Context, uid string) ([]Withdrawal, error)
}

type UIDKey struct{}
