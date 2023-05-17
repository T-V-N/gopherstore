package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strconv"
)

type Accrual struct {
	url string
}

type OrderID struct {
	Order string `json:"order"`
}

type AccrualOrder struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float32 `json:"accrual"`
}

func (a Accrual) RegisterOrder(ctx context.Context, orderID string) error {
	body := bytes.NewBuffer([]byte{})

	err := json.NewEncoder(body).Encode(OrderID{Order: orderID})
	if err != nil {
		return err
	}

	r, err := http.Post(a.url, "application/json", body)

	if err != nil {
		return err
	}

	return r.Body.Close()
}

func (a Accrual) GetOrder(ctx context.Context, orderID string) (AccrualOrder, int, error) {
	r, err := http.Get(a.url + "/api/orders/" + orderID)

	if r.StatusCode == http.StatusTooManyRequests {
		delay, err := strconv.Atoi(r.Header.Get("Retry-After"))
		if err != nil {
			return AccrualOrder{}, 0, err
		}

		return AccrualOrder{}, delay, nil
	}

	if err != nil {
		return AccrualOrder{}, 0, err
	}

	defer r.Body.Close()
	var o AccrualOrder

	err = json.NewDecoder(r.Body).Decode(&o)

	if err != nil {
		return AccrualOrder{}, 0, err
	}

	return o, 0, err
}

func InitAccrual(url string) Accrual {
	return Accrual{url: url}
}
