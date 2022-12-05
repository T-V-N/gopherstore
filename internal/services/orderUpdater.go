package service

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/T-V-N/gopherstore/internal/config"
	sharedTypes "github.com/T-V-N/gopherstore/internal/shared_types"
	"github.com/T-V-N/gopherstore/internal/storage"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Job struct {
	OrderID string
	Status  string
}

type Updater struct {
	JobQueue []Job
	Ch       chan *Job
	cfg      config.Config
	order    sharedTypes.OrderStorage
	user     sharedTypes.UserStorage
	logger   *zap.SugaredLogger
}

type AccrualOrder struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float32 `json:"accrual"`
}

func (u *Updater) checkOrder(orderID, status string) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	r, err := http.Get(u.cfg.AccrualSystemAddress + "/api/orders/" + orderID)

	if err != nil {
		u.logger.Errorw("Error while getting info from accrual service",
			"order id", orderID,
			"accrual address", u.cfg.AccrualSystemAddress,
			"err", err,
		)

		return
	}

	defer r.Body.Close()

	var o AccrualOrder
	err = json.NewDecoder(r.Body).Decode(&o)

	if err != nil {
		u.logger.Errorw("Error while decoding response from accrual service",
			"order id", orderID,
			"accrual address", u.cfg.AccrualSystemAddress,
		)

		return
	}

	if o.Status != status {
		err = u.order.UpdateOrder(ctx, orderID, o.Status, o.Accrual, u.user)
		if err != nil {
			u.logger.Errorw("Error while updating order data",
				"order id", orderID,
				"status", o.Status,
				"uid", u.user,
				"accrual address", u.cfg.AccrualSystemAddress,
				"err", err,
			)
		}
	}
}

func InitUpdater(cfg config.Config, conn *pgxpool.Pool, workerLimit int, logger *zap.SugaredLogger) {
	jobCh := make(chan *Job)

	order, err := storage.InitOrder(conn)

	if err != nil {
		logger.Panicw("Unable to start logger",
			"DB URL", cfg.DatabaseURI,
			"err", err,
		)

		return
	}

	user, err := storage.InitUser(conn)

	if err != nil {
		logger.Panicw("Unable to start logger",
			"DB URL", cfg.DatabaseURI,
			"err", err,
		)

		return
	}

	u := Updater{JobQueue: []Job{}, cfg: cfg, Ch: jobCh, order: order, user: user, logger: logger}

	for i := 0; i < workerLimit; i++ {
		go func() {
			for job := range jobCh {
				u.checkOrder(job.OrderID, job.Status)
			}
		}()
	}

	ticker := time.NewTicker(10 * time.Second)

	for range ticker.C {
		orders, err := order.GetUnproccessedOrders(context.Background())

		if err != nil {
			u.logger.Errorw("Error while getting unprocessed orders",
				"accrual address", u.cfg.AccrualSystemAddress,
				"err", err,
			)
		}

		for _, order := range orders {
			job := &Job{OrderID: order.Number, Status: order.Status}

			u.logger.Infow("new job for order",
				"order id", order.Number,
			)

			jobCh <- job
		}
	}
}
