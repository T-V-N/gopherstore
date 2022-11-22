package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

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
		name string
		body interface{}
		want want
	}{
		{
			name: "Wrong request",
			body: nil,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "Valid login and password",
			body: sharedTypes.Credentials{Login: "tester", Password: "password"},
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name: "Already existing login",
			body: sharedTypes.Credentials{Login: "tester", Password: "password"},
			want: want{
				statusCode: http.StatusConflict,
			},
		},
		{
			name: "Non valid credentials are being rejected",
			body: sharedTypes.Credentials{Login: "t", Password: "p"},
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
			json.NewEncoder(body).Encode(tt.body)
			request := httptest.NewRequest(http.MethodPost, "/", body)
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
		name string
		body interface{}
		want want
	}{
		{
			name: "Wrong request",
			body: nil,
			want: want{
				statusCode: http.StatusBadRequest,
				auth:       "",
			},
		},
		{
			name: "Login with existing creds",
			body: sharedTypes.Credentials{Login: "tester", Password: "password"},
			want: want{
				statusCode: http.StatusOK,
				auth:       "Bearer",
			},
		},
		{
			name: "Login with non-existing creds",
			body: sharedTypes.Credentials{Login: "faker", Password: "tester"},
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

	st.On("GetUser", mock.Anything, mock.Anything).Return("some_uid", nil).Once()
	st.On("GetUser", mock.Anything, mock.Anything).Return("", utils.ErrNotAuthorized)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := bytes.NewBuffer([]byte{})
			json.NewEncoder(body).Encode(tt.body)
			request := httptest.NewRequest(http.MethodPost, "/", body)
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
		name     string
		body     []byte
		want     want
		mockData mockSettings
	}{
		{
			name: "Order already exists",
			body: []byte("12345678903"),
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
			name: "Order created",
			body: []byte("12345678903"),
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
			name: "Bad request",
			body: []byte{},
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
			name: "Already exists",
			body: []byte("12345678903"),
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
			name: "Wrong order id format",
			body: []byte("12345678901"),
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
			w := httptest.NewRecorder()
			hn.HandleCreateOrder(w, request)

			assert.Equal(t, tt.want.statusCode, w.Code)
		})
	}
}

// func Test_HandleListOrder(t *testing.T) {
// 	type want struct {
// 		statusCode int
// 	}

// 	type mockSettings struct {
// 		isNeeded bool
// 		method   string
// 		args     []interface{}
// 		result   interface{}
// 	}

// 	tests := []struct {
// 		name     string
// 		body     []byte
// 		want     want
// 		mockData mockSettings
// 	}{
// 		{
// 			name: "Order already exists",
// 			body: []byte("12345678903"),
// 			want: want{
// 				statusCode: http.StatusOK,
// 			},
// 			mockData: mockSettings{
// 				isNeeded: true,
// 				method:   "CreateOrder",
// 				args:     []interface{}{mock.Anything, "12345678903", mock.Anything},
// 				result: struct {
// 					list []sharedTypes.Order
// 					err  error
// 				}{[]sharedTypes.Order{{Number: "123", Status: "PROCESSING", UploadedAt: "123"}}, nil},
// 			},
// 		},
// 	}
// 	cfg, _ := InitTestConfig()
// 	st := mocks.NewStorage(t)
// 	a := app.InitApp(st, cfg)
// 	hn := handler.InitHandler(a)

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if tt.mockData.isNeeded {
// 				st.On(tt.mockData.method, tt.mockData.args...).Return(tt.mockData.result).Once()
// 			}

// 			request := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(tt.body))
// 			w := httptest.NewRecorder()
// 			hn.HandleCreateOrder(w, request)

// 			assert.Equal(t, tt.want.statusCode, w.Code)
// 		})
// 	}
// }
