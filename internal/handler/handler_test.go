package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/T-V-N/gopherstore/internal/app"
	"github.com/T-V-N/gopherstore/internal/config"
	"github.com/T-V-N/gopherstore/internal/handler"
	"github.com/T-V-N/gopherstore/internal/utils"

	"github.com/T-V-N/gopherstore/mocks"

	sharedTypes "github.com/T-V-N/gopherstore/internal/shared_types"
	"github.com/caarlos0/env/v6"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func InitTestConfig() (*config.Config, error) {
	cfg := &config.Config{}
	err := env.Parse(cfg)

	if err != nil {
		return nil, fmt.Errorf("error: %w", err)
	}

	return cfg, nil
}

func Test_HandlerRegister(t *testing.T) {
	type want struct {
		response   string
		statusCode int
	}

	tests := []struct {
		name        string
		contentType string
		body        interface{}
		want        want
	}{
		{
			name:        "Wrong request",
			contentType: "application/json",
			body:        nil,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:        "Wrong header",
			contentType: "hehehey!",
			body:        nil,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:        "Valid login and password",
			contentType: "application/json",
			body:        sharedTypes.Credentials{Login: "tester", Password: "password"},
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name:        "Already existing login",
			contentType: "application/json",
			body:        sharedTypes.Credentials{Login: "tester", Password: "password"},
			want: want{
				statusCode: http.StatusConflict,
			},
		},
		{
			name:        "Non valid credentials are being rejected",
			contentType: "application/json",
			body:        sharedTypes.Credentials{Login: "t", Password: "p"},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}
	cfg, _ := InitTestConfig()
	st := mocks.NewStorage(t)
	a := app.InitApp(st, cfg)
	hn := handler.InitHandler(a)

	st.On("CreateUser", mock.Anything, mock.Anything).Return("some_uid", nil).Once()
	st.On("CreateUser", mock.Anything, mock.Anything).Return("", utils.ErrDuplicate)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := bytes.NewBuffer([]byte{})
			err := json.NewEncoder(body).Encode(tt.body)

			if err != nil {
				panic("JSON unmarshall error")
			}

			request := httptest.NewRequest(http.MethodPost, "/", body)
			request.Header.Add("Content-type", tt.contentType)
			w := httptest.NewRecorder()
			hn.HandleRegister(w, request)

			assert.Equal(t, tt.want.statusCode, w.Code)
		})
	}
}

func Test_HandlerLogin(t *testing.T) {
	type want struct {
		response   string
		statusCode int
		auth       string
	}

	tests := []struct {
		name        string
		contentType string
		body        interface{}
		want        want
	}{
		{
			name:        "Wrong request",
			contentType: "application/json",
			body:        nil,
			want: want{
				statusCode: http.StatusBadRequest,
				auth:       "",
			},
		},
		{
			name:        "Wrong contentType",
			contentType: "",
			body:        sharedTypes.Credentials{Login: "tester", Password: "password"},
			want: want{
				statusCode: http.StatusBadRequest,
				auth:       "",
			},
		},
		{
			name:        "Login with existing creds",
			contentType: "application/json",
			body:        sharedTypes.Credentials{Login: "tester", Password: "password"},
			want: want{
				statusCode: http.StatusOK,
				auth:       "Bearer",
			},
		},
		{
			name:        "Login with non-existing creds",
			contentType: "application/json",
			body:        sharedTypes.Credentials{Login: "faker", Password: "tester"},
			want: want{
				statusCode: http.StatusUnauthorized,
				auth:       "",
			},
		},
	}
	cfg, _ := InitTestConfig()
	st := mocks.NewStorage(t)
	a := app.InitApp(st, cfg)
	hn := handler.InitHandler(a)

	st.On("GetUser", mock.Anything, mock.Anything).Return(sharedTypes.User{UID: "1", Login: "tester", PasswordHash: "$2a$14$Shj508U123/afnKaPZV4BOTlR3Dt89EGONrff25rbZsg49vzdo8Ga", CurrentBalance: 0, Withdrawn: 0, CreatedAt: "-"}, nil).Once()
	st.On("GetUser", mock.Anything, mock.Anything).Return(sharedTypes.User{}, utils.ErrNotAuthorized)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := bytes.NewBuffer([]byte{})
			err := json.NewEncoder(body).Encode(tt.body)

			if err != nil {
				panic("JSON unmarshall error")
			}

			request := httptest.NewRequest(http.MethodPost, "/", body)
			request.Header.Add("Content-type", tt.contentType)

			w := httptest.NewRecorder()
			hn.HandleLogin(w, request)

			bearerToken := w.Result().Header.Get("Authorization")

			assert.Equal(t, tt.want.statusCode, w.Code)
			assert.Contains(t, bearerToken, tt.want.auth)
		})
	}
}

func Test_RegisterAndLogin(t *testing.T) {
	cfg, _ := InitTestConfig()
	st := mocks.NewStorage(t)
	a := app.InitApp(st, cfg)
	hn := handler.InitHandler(a)

	st.On("CreateUser", mock.Anything, mock.Anything).Return("some_uid", nil)

	t.Run("Check login after register", func(t *testing.T) {
		body := bytes.NewBuffer([]byte{})
		json.NewEncoder(body).Encode(sharedTypes.Credentials{Login: "tester", Password: "password"})
		request := httptest.NewRequest(http.MethodPost, "/", body)
		request.Header.Add("Content-type", "application/json")

		w := httptest.NewRecorder()
		hn.HandleRegister(w, request)

		bearerToken := w.Result().Header.Get("Authorization")

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, bearerToken, "Bearer")
	})
}

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
	st := mocks.NewStorage(t)
	a := app.InitApp(st, cfg)
	hn := handler.InitHandler(a)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockData.isNeeded {
				st.On(tt.mockData.method, tt.mockData.args...).Return(tt.mockData.result).Once()
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
	st := mocks.NewStorage(t)
	a := app.InitApp(st, cfg)
	hn := handler.InitHandler(a)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st.On(tt.mockData.method, tt.mockData.args...).Return(tt.mockData.result...).Once()

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

func Test_HandleGetBalance(t *testing.T) {
	type want struct {
		statusCode   int
		responseBody sharedTypes.Balance
	}

	type mockSettings struct {
		method string
		args   []interface{}
		result []interface{}
	}

	tests := []struct {
		name     string
		uid      string
		want     want
		mockData mockSettings
	}{
		{
			name: "Balance returned",
			uid:  "1337",
			want: want{
				statusCode:   http.StatusOK,
				responseBody: sharedTypes.Balance{Current: 0, Withdrawn: 300},
			},
			mockData: mockSettings{
				method: "GetBalance",
				args:   []interface{}{mock.Anything, "1337"},
				result: []interface{}{sharedTypes.Balance{Current: 0, Withdrawn: 300}, nil},
			},
		},
		{
			name: "Other user balance returned",
			uid:  "1",
			want: want{
				statusCode:   http.StatusOK,
				responseBody: sharedTypes.Balance{Current: 4510, Withdrawn: 300},
			},
			mockData: mockSettings{
				method: "GetBalance",
				args:   []interface{}{mock.Anything, "1"},
				result: []interface{}{sharedTypes.Balance{Current: 4510, Withdrawn: 300}, nil},
			},
		},
	}
	cfg, _ := InitTestConfig()
	st := mocks.NewStorage(t)
	a := app.InitApp(st, cfg)
	hn := handler.InitHandler(a)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st.On(tt.mockData.method, tt.mockData.args...).Return(tt.mockData.result...).Once()

			request := httptest.NewRequest(http.MethodGet, "/", nil)

			ctx := context.WithValue(request.Context(), sharedTypes.UIDKey{}, tt.uid)
			request = request.WithContext(ctx)

			w := httptest.NewRecorder()
			hn.HandleGetBalance(w, request)

			var b sharedTypes.Balance
			json.NewDecoder(w.Body).Decode(&b)

			assert.Equal(t, tt.want.statusCode, w.Code)
			assert.Equal(t, tt.want.responseBody, b)
		})
	}
}

func Test_HandleWithdrawBalance(t *testing.T) {
	type want struct {
		statusCode int
	}

	type mockSettings struct {
		method string
		args   []interface{}
		result []interface{}
	}

	tests := []struct {
		name        string
		uid         string
		contentType string
		want        want
		body        sharedTypes.WtihdrawRequest
		mockData    []mockSettings
	}{
		{
			name:        "Valid withdrawal",
			uid:         "1337",
			contentType: "application/json",
			body:        sharedTypes.WtihdrawRequest{OrderID: "12345678903", Sum: 333},
			want: want{
				statusCode: http.StatusOK,
			},
			mockData: []mockSettings{{
				method: "GetBalance",
				args:   []interface{}{mock.Anything, "1337"},
				result: []interface{}{sharedTypes.Balance{Current: float32(334), Withdrawn: float32(300)}, nil},
			}, {
				method: "WithdrawBalance",
				args:   []interface{}{mock.Anything, "1337", "12345678903", float32(333), float32(1), float32(633)},
				result: []interface{}{nil, nil},
			},
			},
		},
		{
			name:        "Not enough money for withdrawal",
			uid:         "1337",
			contentType: "application/json",
			body:        sharedTypes.WtihdrawRequest{OrderID: "12345678903", Sum: 333},
			want: want{
				statusCode: http.StatusPaymentRequired,
			},
			mockData: []mockSettings{{
				method: "GetBalance",
				args:   []interface{}{mock.Anything, "1337"},
				result: []interface{}{sharedTypes.Balance{Current: float32(332), Withdrawn: float32(333)}, nil},
			},
			},
		},
		{
			name:        "Wrong order number",
			uid:         "1337",
			contentType: "application/json",
			body:        sharedTypes.WtihdrawRequest{OrderID: "12345678902", Sum: 333},
			want: want{
				statusCode: http.StatusUnprocessableEntity,
			},
			mockData: []mockSettings{},
		},
	}
	cfg, _ := InitTestConfig()
	st := mocks.NewStorage(t)
	a := app.InitApp(st, cfg)
	hn := handler.InitHandler(a)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, setting := range tt.mockData {
				st.On(setting.method, setting.args...).Return(setting.result...).Once()
			}

			body := bytes.NewBuffer([]byte{})
			json.NewEncoder(body).Encode(tt.body)

			request := httptest.NewRequest(http.MethodPost, "/", body)
			ctx := context.WithValue(request.Context(), sharedTypes.UIDKey{}, tt.uid)
			request = request.WithContext(ctx)

			request.Header.Add("Content-type", tt.contentType)

			w := httptest.NewRecorder()
			hn.HandleBalanceWithdraw(w, request)

			assert.Equal(t, tt.want.statusCode, w.Code)
		})
	}
}

func Test_HandleListwithdrawal(t *testing.T) {
	type want struct {
		statusCode   int
		responseBody []sharedTypes.Withdrawal
	}

	type mockSettings struct {
		method string
		args   []interface{}
		result []interface{}
	}

	mockWithdrawalList := []sharedTypes.Withdrawal{
		{OrderID: "1", Sum: float32(319), ProcessedAt: "sometime"},
		{OrderID: "2", Sum: float32(13), ProcessedAt: "sometime"},
		{OrderID: "133", Sum: float32(319), ProcessedAt: "sometime"},
	}

	tests := []struct {
		name     string
		uid      string
		want     want
		mockData mockSettings
	}{
		{
			name: "Some withdrawals returned",
			uid:  "1337",
			want: want{
				statusCode:   http.StatusOK,
				responseBody: mockWithdrawalList,
			},
			mockData: mockSettings{
				method: "ListWithdrawals",
				args:   []interface{}{mock.Anything, "1337"},
				result: []interface{}{mockWithdrawalList, nil},
			},
		},
		{
			name: "Empty List",
			uid:  "1337",
			want: want{
				statusCode: http.StatusNoContent,
			},
			mockData: mockSettings{
				method: "ListWithdrawals",
				args:   []interface{}{mock.Anything, "1337"},
				result: []interface{}{[]sharedTypes.Withdrawal{}, nil},
			},
		},
	}
	cfg, _ := InitTestConfig()
	st := mocks.NewStorage(t)
	a := app.InitApp(st, cfg)
	hn := handler.InitHandler(a)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st.On(tt.mockData.method, tt.mockData.args...).Return(tt.mockData.result...).Once()

			request := httptest.NewRequest(http.MethodGet, "/", nil)

			ctx := context.WithValue(request.Context(), sharedTypes.UIDKey{}, tt.uid)
			request = request.WithContext(ctx)

			w := httptest.NewRecorder()
			hn.HandleListWithdrawals(w, request)

			var l []sharedTypes.Withdrawal
			json.NewDecoder(w.Body).Decode(&l)

			assert.Equal(t, tt.want.statusCode, w.Code)
			assert.Equal(t, len(tt.want.responseBody), len(l))
		})
	}
}
