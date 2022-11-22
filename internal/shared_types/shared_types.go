package sharedTypes

import (
	"context"
)

type Credentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Order struct {
	Number     string  `json:"number"`
	Status     string  `json:"status"`
	Accural    float32 `json:"accural,omitempty"`
	UploadedAt string  `json:"uploaded_at"`
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
	OrderID      string  `json:"order"`
	Sum          float32 `json:"sum"`
	ProccessedAt string  `json:"processed_at"`
}

type User struct {
	UID            string
	Login          string
	PasswordHash   string
	CurrentBalance float32
	Withdrawn      float32
	CreatedAt      string
}

type Storage interface {
	CreateUser(ctx context.Context, creds Credentials) (string, error)
	GetUser(ctx context.Context, creds Credentials) (User, error)
	CreateOrder(ctx context.Context, orderID, uid string) error
	ListOrders(ctx context.Context, uid string) ([]Order, error)
	GetBalance(ctx context.Context, uid string) (Balance, error)
	WithdrawBalance(ctx context.Context, uid, orderID string, amount float32) error
	GetListWithdrawals(ctx context.Context, uid string) ([]Withdrawal, error)
}

type UIDKey struct{}
