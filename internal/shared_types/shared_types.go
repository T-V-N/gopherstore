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

type UserStorage interface {
	CreateUser(context.Context, Credentials) (string, error)
	GetUser(context.Context, Credentials) (User, error)
	GetBalance(context.Context, string) (Balance, error)
	GetBalanceAndLock(context.Context, string) (Balance, error)
	WithdrawBalance(context.Context, string, string, float32, float32, float32, WithdrawalStorage) error
	UpdateUser(context.Context, string, float32) error
}

type OrderStorage interface {
	CreateOrder(context.Context, string, string) error
	ListOrders(context.Context, string) ([]Order, error)
	GetUnproccessedOrders(context.Context) ([]Order, error)
	UpdateOrder(context.Context, string, string, float32, UserStorage) error
}

type WithdrawalStorage interface {
	ListWithdrawals(context.Context, string) ([]Withdrawal, error)
	CreateWithdrawal(context.Context, string, float32, string) error
}

type UIDKey struct{}
