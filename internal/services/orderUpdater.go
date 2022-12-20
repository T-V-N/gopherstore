package service

import (
	"context"
	"sync"
	"time"

	"github.com/T-V-N/gopherstore/internal/config"
	sharedTypes "github.com/T-V-N/gopherstore/internal/shared_types"
	"github.com/T-V-N/gopherstore/internal/utils"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Job struct {
	OrderID string
	Status  string
}

func checkOrder(orderID, status string, logger *zap.SugaredLogger, cfg config.Config, ch chan *Job, user sharedTypes.UserApper, order sharedTypes.OrderApper, accrual utils.Accrual) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.CheckOrderDelay)*time.Second)
	defer cancel()

	o, delay, err := accrual.GetOrder(ctx, orderID)

	if err != nil {
		logger.Errorw("Error while decoding response from accrual service",
			"order id", orderID,
			"accrual address", cfg.AccrualSystemAddress,
		)

		return
	}

	if delay != 0 {
		job := &Job{OrderID: orderID, Status: status}

		logger.Infow("job delayed for some time",
			"order id", orderID,
		)

		time.Sleep(time.Duration(delay) * time.Second)
		ch <- job

		return
	}

	if o.Status != status {
		err = order.UpdateOrder(ctx, orderID, o.Status, o.Accrual, user)
		if err != nil {
			logger.Errorw("Error while updating order data",
				"order id", orderID,
				"status", o.Status,
				"uid", user,
				"accrual address", cfg.AccrualSystemAddress,
				"err", err,
			)
		}
	}
}

func InitUpdater(ctx context.Context, cfg config.Config, conn *pgxpool.Pool, workerLimit int, logger *zap.SugaredLogger, User sharedTypes.UserApper, Order sharedTypes.OrderApper) {
	jobCh := make(chan *Job)
	wg := sync.WaitGroup{}
	accrual := utils.InitAccrual(cfg.AccrualSystemAddress + "/api/orders")

	for i := 0; i < workerLimit; i++ {
		wg.Add(1)

		go func() {
			for {
				select {
				case job := <-jobCh:
					checkOrder(job.OrderID, job.Status, logger, cfg, jobCh, User, Order, accrual)
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
				logger.Errorw("Error while getting unprocessed orders",
					"accrual address", cfg.AccrualSystemAddress,
					"err", err,
				)
			}

			for _, order := range orders {
				job := &Job{OrderID: order.Number, Status: order.Status}

				logger.Infow("new job for order",
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
