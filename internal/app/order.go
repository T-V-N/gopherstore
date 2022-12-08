package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/T-V-N/gopherstore/internal/config"
	sharedTypes "github.com/T-V-N/gopherstore/internal/shared_types"
	"github.com/T-V-N/gopherstore/internal/storage"
	"github.com/T-V-N/gopherstore/internal/utils"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joeljunstrom/go-luhn"
	"go.uber.org/zap"
)

type OrderApp struct {
	Order  sharedTypes.OrderStorage
	Cfg    *config.Config
	logger *zap.SugaredLogger
}

type OrderID struct {
	Order string `json:"order"`
}

func InitOrderApp(Conn *pgxpool.Pool, cfg *config.Config, logger *zap.SugaredLogger) (*OrderApp, error) {
	order, err := storage.InitOrder(Conn)

	if err != nil {
		return nil, err
	}

	return &OrderApp{order, cfg, logger}, nil
}

func (app *OrderApp) CreateOrder(ctx context.Context, orderID string, uid string) error {
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
