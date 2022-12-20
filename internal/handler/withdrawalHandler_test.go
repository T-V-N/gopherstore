package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/T-V-N/gopherstore/internal/app"
	"github.com/T-V-N/gopherstore/internal/handler"
	"go.uber.org/zap"

	"github.com/T-V-N/gopherstore/mocks"

	sharedTypes "github.com/T-V-N/gopherstore/internal/shared_types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_HandleWithdrawBalance(t *testing.T) {
	type want struct {
		statusCode int
	}

	type mockSettings struct {
		storageTyp string
		method     string
		args       []interface{}
		result     []interface{}
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
				storageTyp: "user",
				method:     "GetBalance",
				args:       []interface{}{mock.Anything, "1337"},
				result:     []interface{}{sharedTypes.Balance{Current: float32(334), Withdrawn: float32(300)}, nil},
			}, {
				storageTyp: "withdrawal",
				method:     "WithdrawBalance",
				args:       []interface{}{mock.Anything, "1337", "12345678903", float32(333), float32(1), float32(633)},
				result:     []interface{}{nil, nil},
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
				storageTyp: "user",
				method:     "GetBalance",
				args:       []interface{}{mock.Anything, "1337"},
				result:     []interface{}{sharedTypes.Balance{Current: float32(332), Withdrawn: float32(333)}, nil},
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

	user := mocks.NewUserStorage(t)
	withdrawal := mocks.NewWithdrawalStorage(t)

	a := app.UserApp{User: user, Withdrawal: withdrawal, Cfg: cfg, UserLocks: &sync.Map{}}

	hn := handler.InitUserHandler(&a, cfg, &zap.SugaredLogger{})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, setting := range tt.mockData {
				switch setting.storageTyp {
				case "user":
					user.On(setting.method, setting.args...).Return(setting.result...).Once()
				case "withdrawal":
					withdrawal.On(setting.method, setting.args...).Return(setting.result...).Once()
				}
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

	mockTime := time.Now()
	mockWithdrawalList := []sharedTypes.Withdrawal{
		{ID: "1", Sum: float32(319), ProcessedAt: mockTime},
		{ID: "2", Sum: float32(13), ProcessedAt: mockTime},
		{ID: "133", Sum: float32(319), ProcessedAt: mockTime},
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
	withdrawal := mocks.NewWithdrawalStorage(t)

	a := app.WithdrawalApp{Withdrawal: withdrawal, Cfg: cfg}
	hn := handler.InitWithdrawalHandler(&a, cfg, &zap.SugaredLogger{})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withdrawal.On(tt.mockData.method, tt.mockData.args...).Return(tt.mockData.result...).Once()

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
