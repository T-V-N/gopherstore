package service

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/T-V-N/gopherstore/internal/config"
	sharedTypes "github.com/T-V-N/gopherstore/internal/shared_types"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Job struct {
	OrderID string
	Status  string
}

type Updater struct {
	Ch              chan *Job
	cfg             config.Config
	order           sharedTypes.OrderApper
	user            sharedTypes.UserApper
	logger          *zap.SugaredLogger
	done            chan bool
	CheckOrderDelay uint
}

type AccrualOrder struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float32 `json:"accrual"`
}

func (u *Updater) checkOrder(orderID, status string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(u.CheckOrderDelay)*time.Second)
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

	if r.StatusCode == http.StatusTooManyRequests {
		delay, err := strconv.Atoi(r.Header.Get("Retry-After"))
		if err != nil {
			u.logger.Infow("Unable to parse retry-after header")
			return
		}

		go func() {
			job := &Job{OrderID: orderID, Status: status}

			u.logger.Infow("job delayed by 60 sec",
				"order id", orderID,
			)

			time.Sleep(time.Duration(delay) * time.Second)
			u.Ch <- job
		}()

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

func InitUpdater(ctx context.Context, cfg config.Config, conn *pgxpool.Pool, workerLimit int, logger *zap.SugaredLogger, User sharedTypes.UserApper, Order sharedTypes.OrderApper) {
	jobCh := make(chan *Job)

	u := Updater{cfg: cfg, Ch: jobCh, order: Order, user: User, logger: logger, CheckOrderDelay: cfg.CheckOrderDelay}

	wg := sync.WaitGroup{}

	for i := 0; i < workerLimit; i++ {
		wg.Add(1)

		go func() {
			for {
				select {
				case job := <-jobCh:
					u.checkOrder(job.OrderID, job.Status)
				case <-ctx.Done():
					wg.Done()
					return
				}
			}
		}()
	}

	ticker := time.NewTicker(time.Duration(cfg.CheckOrderInterval) * time.Second)

	for {
		select {
		case <-ticker.C:
			requestCtx, stop := context.WithTimeout(ctx, time.Duration(cfg.CheckOrderDelay)*time.Second)
			defer stop()

			orders, err := Order.GetUnproccessedOrders(requestCtx)

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
		case <-ctx.Done():
			close(jobCh)
			wg.Wait()
			logger.Info("Workers gracefully stopped")

			return
		}
	}
}
