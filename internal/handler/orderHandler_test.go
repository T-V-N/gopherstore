package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/T-V-N/gopherstore/internal/app"
	"github.com/T-V-N/gopherstore/internal/handler"
	"github.com/T-V-N/gopherstore/internal/utils"
	"go.uber.org/zap"

	"github.com/T-V-N/gopherstore/mocks"

	sharedTypes "github.com/T-V-N/gopherstore/internal/shared_types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_HandleCreateOrder(t *testing.T) {
	type want struct {
		statusCode int
	}

	type mockSettings struct {
		isNeeded bool
		method   string
		args     []interface{}
		result   interface{}
	}

	tests := []struct {
		name        string
		body        []byte
		contentType string
		want        want
		mockData    mockSettings
	}{
		{
			name:        "Order already exists",
			body:        []byte("12345678903"),
			contentType: "text/plain",
			want: want{
				statusCode: http.StatusOK,
			},
			mockData: mockSettings{
				isNeeded: true,
				method:   "CreateOrder",
				args:     []interface{}{mock.Anything, "12345678903", mock.Anything},
				result:   utils.ErrAlreadyCreated,
			},
		},
		{
			name:        "Order created",
			body:        []byte("12345678903"),
			contentType: "text/plain",
			want: want{
				statusCode: http.StatusAccepted,
			},
			mockData: mockSettings{
				isNeeded: true,
				method:   "CreateOrder",
				args:     []interface{}{mock.Anything, "12345678903", mock.Anything},
				result:   nil,
			},
		},
		{
			name:        "Bad request",
			body:        []byte{},
			contentType: "text/plain",
			want: want{
				statusCode: http.StatusBadRequest,
			},
			mockData: mockSettings{
				isNeeded: false,
				method:   "CreateOrder",
				args:     []interface{}{mock.Anything, mock.Anything, mock.Anything},
				result:   mock.Anything,
			},
		},
		{
			name:        "Already exists",
			body:        []byte("12345678903"),
			contentType: "text/plain",
			want: want{
				statusCode: http.StatusConflict,
			},
			mockData: mockSettings{
				isNeeded: true,
				method:   "CreateOrder",
				args:     []interface{}{mock.Anything, "12345678903", mock.Anything},
				result:   utils.ErrDuplicate,
			},
		},
		{
			name:        "Wrong order id format",
			body:        []byte("12345678901"),
			contentType: "text/plain",
			want: want{
				statusCode: http.StatusUnprocessableEntity,
			},
			mockData: mockSettings{
				isNeeded: false,
				method:   "CreateOrder",
				args:     []interface{}{mock.Anything, mock.Anything, mock.Anything},
				result:   mock.Anything,
			},
		},
	}
	cfg, _ := InitTestConfig()
	order := mocks.NewOrderStorager(t)

	a := app.OrderApp{Order: order, Cfg: cfg, RegOrder: utils.InitAccrual(cfg.AccrualSystemAddress + "/api/orders")}
	hn := handler.InitOrderHandler(&a, cfg, &zap.SugaredLogger{})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockData.isNeeded {
				order.On(tt.mockData.method, tt.mockData.args...).Return(tt.mockData.result).Once()
			}

			request := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(tt.body))
			request.Header.Add("Content-type", tt.contentType)

			w := httptest.NewRecorder()
			hn.HandleCreateOrder(w, request)

			assert.Equal(t, tt.want.statusCode, w.Code)
		})
	}
}

func Test_HandleListOrder(t *testing.T) {
	type want struct {
		statusCode   int
		responseBody []sharedTypes.Order
	}

	type mockSettings struct {
		method string
		args   []interface{}
		result []interface{}
	}

	mockTime := time.Now()
	mockOrderList := []sharedTypes.Order{
		{Number: "1", Status: "NEW", Accrual: 0, UploadedAt: mockTime},
		{Number: "2", Status: "NEW", Accrual: 33, UploadedAt: mockTime},
		{Number: "133", Status: "INVALID", Accrual: 0, UploadedAt: mockTime},
	}

	tests := []struct {
		name     string
		uid      string
		want     want
		mockData mockSettings
	}{
		{
			name: "Some orders returned",
			uid:  "1337",
			want: want{
				statusCode:   http.StatusOK,
				responseBody: mockOrderList,
			},
			mockData: mockSettings{
				method: "ListOrders",
				args:   []interface{}{mock.Anything, "1337"},
				result: []interface{}{mockOrderList, nil},
			},
		},
		{
			name: "Empty List",
			uid:  "1337",
			want: want{
				statusCode: http.StatusNoContent,
			},
			mockData: mockSettings{
				method: "ListOrders",
				args:   []interface{}{mock.Anything, "1337"},
				result: []interface{}{[]sharedTypes.Order{}, nil},
			},
		},
	}
	cfg, _ := InitTestConfig()
	order := mocks.NewOrderStorager(t)

	a := app.OrderApp{Order: order, Cfg: cfg}
	hn := handler.InitOrderHandler(&a, cfg, &zap.SugaredLogger{})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order.On(tt.mockData.method, tt.mockData.args...).Return(tt.mockData.result...).Once()

			request := httptest.NewRequest(http.MethodGet, "/", nil)

			ctx := context.WithValue(request.Context(), sharedTypes.UIDKey{}, tt.uid)
			request = request.WithContext(ctx)

			w := httptest.NewRecorder()
			hn.HandleListOrder(w, request)

			var l []sharedTypes.Order
			json.NewDecoder(w.Body).Decode(&l)

			assert.Equal(t, tt.want.statusCode, w.Code)
			assert.Equal(t, len(tt.want.responseBody), len(l))
		})
	}
}
