package app

import (
	"context"

	"github.com/T-V-N/gopherstore/internal/config"
	sharedTypes "github.com/T-V-N/gopherstore/internal/shared_types"
	"github.com/T-V-N/gopherstore/internal/storage"
	"github.com/T-V-N/gopherstore/internal/utils"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joeljunstrom/go-luhn"
	"go.uber.org/zap"
)

type OrderApp struct {
	Order    sharedTypes.OrderStorager
	Cfg      *config.Config
	logger   *zap.SugaredLogger
	RegOrder sharedTypes.OrderRegisterer
}

func InitOrderApp(Conn *pgxpool.Pool, cfg *config.Config, logger *zap.SugaredLogger, or sharedTypes.OrderRegisterer) (*OrderApp, error) {
	order, err := storage.InitOrder(Conn)

	if err != nil {
		return nil, err
	}

	return &OrderApp{order, cfg, logger, or}, nil
}

func (app *OrderApp) CreateOrder(ctx context.Context, orderID string, uid string) error {
	isOrderIDValid := luhn.Valid(orderID)

	if !isOrderIDValid {
		return utils.ErrWrongFormat
	}

	err := app.RegOrder.RegisterOrder(ctx, orderID)

	if err != nil {
		return err
	}

	err = app.Order.CreateOrder(ctx, orderID, uid)

	return err
}

func (app *OrderApp) ListOrders(ctx context.Context, uid string) ([]sharedTypes.Order, error) {
	list, err := app.Order.ListOrders(ctx, uid)

	if err != nil {
		return nil, err
	}

	if len(list) == 0 {
		return nil, utils.ErrNoData
	}

	return list, err
}

func (app *OrderApp) GetUnproccessedOrders(ctx context.Context) ([]sharedTypes.Order, error) {
	list, err := app.Order.GetUnproccessedOrders(ctx)

	if err != nil {
		return nil, err
	}

	if len(list) == 0 {
		return nil, utils.ErrNoData
	}

	return list, err
}

func (app *OrderApp) UpdateOrder(ctx context.Context, orderID, status string, accrual float32, user sharedTypes.UserApper) error {
	uid, err := app.Order.UpdateOrder(ctx, orderID, status, accrual)

	if err != nil {
		return err
	}

	if accrual > 0 {
		return user.UpdateUser(ctx, uid, orderID, accrual)
	}

	return nil
}
