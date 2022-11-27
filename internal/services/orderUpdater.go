package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/T-V-N/gopherstore/internal/config"
	"github.com/T-V-N/gopherstore/internal/storage"
)

type Job struct {
	OrderID string
	Status  string
}

type Updater struct {
	JobQueue []Job
	Ch       chan *Job
	cfg      config.Config
	st       storage.Storage
}

type AccuralOrder struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accural float32 `json:"accural"`
}

func (u *Updater) checkOrder(orderID, status string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	r, err := http.Get(u.cfg.AccuralSystemAddress + "/api/orders/" + orderID)

	if err != nil {
		fmt.Print(err)
		return
	}

	var o AccuralOrder
	err = json.NewDecoder(r.Body).Decode(&o)

	if err != nil {
		fmt.Print(err)
		return
	}

	if o.Status != status {
		err = u.st.UpdateOrder(ctx, orderID, o.Status, o.Accural)
		if err != nil {
			fmt.Print(err)
		}
	}
}

func InitUpdater(cfg config.Config, st storage.Storage, workerLimit int) {
	jobCh := make(chan *Job)
	u := Updater{JobQueue: []Job{}, cfg: cfg, Ch: jobCh, st: st}
	for i := 0; i < workerLimit; i++ {
		go func() {
			for job := range jobCh {
				u.checkOrder(job.OrderID, job.Status)
			}
		}()
	}

	ticker := time.NewTicker(10 * time.Second)

	for _ = range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
		defer cancel()

		orders, err := st.GetUnproccessedOrders(ctx)

		if err != nil {
			fmt.Print(err)
		}

		for _, order := range orders {
			job := &Job{OrderID: order.Number, Status: order.Status}
			jobCh <- job
		}
	}
}
