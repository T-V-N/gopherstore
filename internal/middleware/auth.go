package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/T-V-N/gopherstore/internal/auth"
	"github.com/T-V-N/gopherstore/internal/config"
	sharedTypes "github.com/T-V-N/gopherstore/internal/shared_types"
	"github.com/T-V-N/gopherstore/internal/utils"
)

func InitAuth(cfg *config.Config) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")

			if token == "" {
				http.Error(w, utils.ErrNotAuthorized.Error(), http.StatusUnauthorized)
				return
			}

			headerParts := strings.Split(token, " ")
			if len(headerParts) != 2 || headerParts[0] != "Bearer" {
				http.Error(w, utils.ErrNotAuthorized.Error(), http.StatusUnauthorized)
				return
			}

			uid, err := auth.ParseToken(headerParts[1], []byte(cfg.SecretKey))

			if err != nil || uid == "" {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), sharedTypes.UIDKey{}, uid)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
