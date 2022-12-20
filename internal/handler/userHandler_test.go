package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/T-V-N/gopherstore/internal/app"
	"github.com/T-V-N/gopherstore/internal/handler"
	"github.com/T-V-N/gopherstore/internal/utils"
	"go.uber.org/zap"

	"github.com/T-V-N/gopherstore/mocks"

	sharedTypes "github.com/T-V-N/gopherstore/internal/shared_types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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
	user := mocks.NewUserStorage(t)
	withdrawal := mocks.NewWithdrawalStorage(t)

	a := app.UserApp{User: user, Withdrawal: withdrawal, Cfg: cfg}
	hn := handler.InitUserHandler(&a, cfg, &zap.SugaredLogger{})

	user.On("CreateUser", mock.Anything, mock.Anything).Return("some_uid", nil).Once()
	user.On("CreateUser", mock.Anything, mock.Anything).Return("", utils.ErrDuplicate)

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
	user := mocks.NewUserStorage(t)
	withdrawal := mocks.NewWithdrawalStorage(t)

	a := app.UserApp{User: user, Withdrawal: withdrawal, Cfg: cfg}
	hn := handler.InitUserHandler(&a, cfg, &zap.SugaredLogger{})

	user.On("GetUser", mock.Anything, mock.Anything).Return(sharedTypes.User{UID: "1", Login: "tester", PasswordHash: "$2a$14$Shj508U123/afnKaPZV4BOTlR3Dt89EGONrff25rbZsg49vzdo8Ga", CurrentBalance: 0, Withdrawn: 0, CreatedAt: "-"}, nil).Once()
	user.On("GetUser", mock.Anything, mock.Anything).Return(sharedTypes.User{}, utils.ErrNotAuthorized)

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

			res := w.Result()
			res.Body.Close()

			bearerToken := res.Header.Get("Authorization")

			assert.Equal(t, tt.want.statusCode, w.Code)
			assert.Contains(t, bearerToken, tt.want.auth)
		})
	}
}

func Test_RegisterAndLogin(t *testing.T) {
	cfg, _ := InitTestConfig()

	user := mocks.NewUserStorage(t)
	withdrawal := mocks.NewWithdrawalStorage(t)

	a := app.UserApp{User: user, Withdrawal: withdrawal, Cfg: cfg}
	hn := handler.InitUserHandler(&a, cfg, &zap.SugaredLogger{})
	user.On("CreateUser", mock.Anything, mock.Anything).Return("some_uid", nil)

	t.Run("Check login after register", func(t *testing.T) {
		body := bytes.NewBuffer([]byte{})
		json.NewEncoder(body).Encode(sharedTypes.Credentials{Login: "tester", Password: "password"})
		request := httptest.NewRequest(http.MethodPost, "/", body)
		request.Header.Add("Content-type", "application/json")

		w := httptest.NewRecorder()
		hn.HandleRegister(w, request)

		res := w.Result()
		res.Body.Close()

		bearerToken := res.Header.Get("Authorization")

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, bearerToken, "Bearer")
	})
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

	user := mocks.NewUserStorage(t)
	withdrawal := mocks.NewWithdrawalStorage(t)

	a := app.UserApp{User: user, Withdrawal: withdrawal, Cfg: cfg}
	hn := handler.InitUserHandler(&a, cfg, &zap.SugaredLogger{})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user.On(tt.mockData.method, tt.mockData.args...).Return(tt.mockData.result...).Once()

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
