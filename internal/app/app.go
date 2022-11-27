package app

import (
	"context"

	"github.com/T-V-N/gopherstore/internal/auth"
	"github.com/T-V-N/gopherstore/internal/config"
	sharedTypes "github.com/T-V-N/gopherstore/internal/shared_types"
	"github.com/T-V-N/gopherstore/internal/utils"
	"github.com/joeljunstrom/go-luhn"
)

type App struct {
	st  sharedTypes.Storage
	Cfg *config.Config
}

func InitApp(st sharedTypes.Storage, cfg *config.Config) *App {
	return &App{st, cfg}
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

	uid, err := app.st.CreateUser(ctx, creds)

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

	user, err := app.st.GetUser(ctx, creds)

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

	err := app.st.CreateOrder(ctx, orderID, uid)

	return err
}

func (app *App) ListOrders(ctx context.Context, uid string) ([]sharedTypes.Order, error) {
	list, err := app.st.ListOrders(ctx, uid)

	if err != nil {
		return nil, err
	}

	if len(list) == 0 {
		return nil, utils.ErrNoData
	}

	return list, err
}

func (app *App) GetBalance(ctx context.Context, uid string) (sharedTypes.Balance, error) {
	balance, err := app.st.GetBalance(ctx, uid)

	return balance, err
}

func (app *App) WithdrawBalance(ctx context.Context, uid, orderID string, amount float32) error {
	isOrderIDValid := luhn.Valid(orderID)

	if !isOrderIDValid {
		return utils.ErrWrongFormat
	}

	balance, err := app.st.GetBalance(ctx, uid)

	if err != nil {
		return err
	}

	if balance.Current-amount < 0 {
		return utils.ErrPaymentError
	}

	newWithdrawn := balance.Withdrawn + amount
	newCurrent := balance.Current - amount

	err = app.st.WithdrawBalance(ctx, uid, orderID, amount, newCurrent, newWithdrawn)

	return err
}

func (app *App) GetListWithdrawals(ctx context.Context, uid string) ([]sharedTypes.Withdrawal, error) {
	list, err := app.st.ListWithdrawals(ctx, uid)

	if len(list) == 0 {
		return []sharedTypes.Withdrawal{}, utils.ErrNoData
	}

	return list, err
}
