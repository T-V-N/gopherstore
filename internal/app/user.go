package app

import (
	"context"

	"github.com/T-V-N/gopherstore/internal/auth"
	"github.com/T-V-N/gopherstore/internal/config"
	sharedTypes "github.com/T-V-N/gopherstore/internal/shared_types"
	"github.com/T-V-N/gopherstore/internal/storage"
	"github.com/T-V-N/gopherstore/internal/utils"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joeljunstrom/go-luhn"
	"go.uber.org/zap"
)

type UserApp struct {
	User       sharedTypes.UserStorager
	Withdrawal sharedTypes.WithdrawalStorager
	Cfg        *config.Config
	logger     *zap.SugaredLogger
}

func InitUserApp(Conn *pgxpool.Pool, w WithdrawalApp, cfg *config.Config, logger *zap.SugaredLogger) (*UserApp, error) {
	user, err := storage.InitUser(Conn)

	if err != nil {
		return nil, err
	}

	return &UserApp{user, w.Withdrawal, cfg, logger}, nil
}

func (app *UserApp) Register(ctx context.Context, creds sharedTypes.Credentials) (string, error) {
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

func (app *UserApp) Login(ctx context.Context, creds sharedTypes.Credentials) (string, error) {
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

func (app *UserApp) GetBalance(ctx context.Context, uid string) (sharedTypes.Balance, error) {
	balance, err := app.User.GetBalance(ctx, uid)

	return balance, err
}

func (app *UserApp) WithdrawBalance(ctx context.Context, uid, orderID string, amount float32) error {
	isOrderIDValid := luhn.Valid(orderID)

	if !isOrderIDValid {
		return utils.ErrWrongFormat
	}

	balance, err := app.User.GetBalanceAndLock(ctx, uid)

	if err != nil {
		return err
	}

	if balance.Current-amount < 0 {
		return utils.ErrPaymentError
	}

	newWithdrawn := balance.Withdrawn + amount
	newCurrent := balance.Current - amount

	err = app.User.WithdrawBalance(ctx, uid, orderID, amount, newCurrent, newWithdrawn, app.Withdrawal)

	return err
}
