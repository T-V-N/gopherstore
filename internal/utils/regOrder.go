package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

type RegOrderHTTP struct {
	url string
}

type OrderID struct {
	Order string `json:"order"`
}

func (ro RegOrderHTTP) RegisterOrder(ctx context.Context, orderID string) error {
	body := bytes.NewBuffer([]byte{})

	err := json.NewEncoder(body).Encode(OrderID{Order: orderID})
	if err != nil {
		return err
	}

	r, err := http.Post(ro.url, "application/json", body)

	if err != nil {
		return err
	}

	return r.Body.Close()
}

func RegOrderHTTPInit(url string) RegOrderHTTP {
	return RegOrderHTTP{url: url}
}
