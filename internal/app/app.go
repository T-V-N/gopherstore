package app

import (
	"context"
	"strconv"

	"github.com/T-V-N/gopherstore/internal/auth"
	"github.com/T-V-N/gopherstore/internal/config"
	sharedTypes "github.com/T-V-N/gopherstore/internal/shared_types"
	"github.com/T-V-N/gopherstore/internal/utils"
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
	ID, err := strconv.Atoi(orderID)
	orderIDValid := utils.Valid(ID)

	if !orderIDValid || err != nil {
		return utils.ErrWrongFormat
	}

	err = app.st.CreateOrder(ctx, orderID, uid)

	return err
}

func (app *App) ListOrders(ctx context.Context, uid string) ([]sharedTypes.Order, error) {
	list, err := app.st.ListOrders(ctx, uid)

	return list, err
}

func (app *App) GetBalance(ctx context.Context, uid string) (sharedTypes.Balance, error) {
	balance, err := app.st.GetBalance(ctx, uid)

	return balance, err
}

func (app *App) WithdrawBalance(ctx context.Context, uid, orderID string, amount float32) error {
	err := app.st.WithdrawBalance(ctx, uid, orderID, amount)

	return err
}

func (app *App) GetListWithdrawals(ctx context.Context, uid string) ([]sharedTypes.Withdrawal, error) {
	list, err := app.st.GetListWithdrawals(ctx, uid)

	return list, err
}
