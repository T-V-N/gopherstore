package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/T-V-N/gopherstore/internal/auth"
	"github.com/T-V-N/gopherstore/internal/config"
	sharedTypes "github.com/T-V-N/gopherstore/internal/shared_types"
	"github.com/T-V-N/gopherstore/internal/storage"
	"github.com/T-V-N/gopherstore/internal/utils"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joeljunstrom/go-luhn"
)

type App struct {
	User       sharedTypes.UserStorage
	Order      sharedTypes.OrderStorage
	Withdrawal sharedTypes.WithdrawalStorage

	Cfg *config.Config
}

type OrderID struct {
	Order string `json:"order"`
}

func InitApp(Conn *pgxpool.Pool, cfg *config.Config) (*App, error) {
	user, err := storage.InitUser(Conn)

	if err != nil {
		return nil, err
	}

	order, err := storage.InitOrder(Conn)

	if err != nil {
		return nil, err
	}

	withdrawal, err := storage.InitWithdrawal(Conn)

	if err != nil {
		return nil, err
	}

	return &App{user, order, withdrawal, cfg}, nil
}

func (app *App) Register(ctx context.Context, creds sharedTypes.Credentials) (string, error) {
	err := utils.ValidateLogPass(creds)

	if err != nil {
		return "", err
	}

	creds.Password, err = utils.HashPassword(creds.Password)

	if err != nil {
		return "", err
	}

	uid, err := app.User.CreateUser(ctx, creds)

	if err != nil {
		return "", err
	}

	return uid, nil
}

func (app *App) Login(ctx context.Context, creds sharedTypes.Credentials) (string, error) {
	err := utils.ValidateLogPass(creds)

	if err != nil {
		return "", err
	}

	user, err := app.User.GetUser(ctx, creds)

	if err != nil {
		return "", utils.ErrNotAuthorized
	}

	isPasswordValid := utils.CheckPasswordHash(creds.Password, user.PasswordHash)

	if !isPasswordValid {
		return "", utils.ErrNotAuthorized
	}

	return auth.CreateToken(user.UID, app.Cfg)
}

func (app *App) CreateOrder(ctx context.Context, orderID string, uid string) error {
	isOrderIDValid := luhn.Valid(orderID)

	if !isOrderIDValid {
		return utils.ErrWrongFormat
	}

	body := bytes.NewBuffer([]byte{})

	err := json.NewEncoder(body).Encode(OrderID{Order: orderID})
	if err != nil {
		return err
	}

	r, err := http.Post(app.Cfg.AccrualSystemAddress+"/api/orders", "application/json", body)

	if err != nil {
		return err
	}

	r.Body.Close()

	err = app.Order.CreateOrder(ctx, orderID, uid)

	return err
}

func (app *App) ListOrders(ctx context.Context, uid string) ([]sharedTypes.Order, error) {
	list, err := app.Order.ListOrders(ctx, uid)

	if err != nil {
		return nil, err
	}

	if len(list) == 0 {
		return nil, utils.ErrNoData
	}

	return list, err
}

func (app *App) GetBalance(ctx context.Context, uid string) (sharedTypes.Balance, error) {
	balance, err := app.User.GetBalance(ctx, uid)

	return balance, err
}

func (app *App) WithdrawBalance(ctx context.Context, uid, orderID string, amount float32) error {
	isOrderIDValid := luhn.Valid(orderID)

	if !isOrderIDValid {
		return utils.ErrWrongFormat
	}

	balance, err := app.User.GetBalance(ctx, uid)

	if err != nil {
		return err
	}

	if balance.Current-amount < 0 {
		return utils.ErrPaymentError
	}

	newWithdrawn := balance.Withdrawn + amount
	newCurrent := balance.Current - amount

	err = app.Withdrawal.WithdrawBalance(ctx, uid, orderID, amount, newCurrent, newWithdrawn)

	return err
}

func (app *App) GetListWithdrawals(ctx context.Context, uid string) ([]sharedTypes.Withdrawal, error) {
	list, err := app.Withdrawal.ListWithdrawals(ctx, uid)

	if len(list) == 0 {
		return []sharedTypes.Withdrawal{}, utils.ErrNoData
	}

	return list, err
}
