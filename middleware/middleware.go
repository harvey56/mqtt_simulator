package middleware

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/gorilla/mux"
)

// AppRouterInjector is a middleware that injects the AppRouter into the request context.
func AppRouterInjector(ar *AppRouter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "appRouter", ar)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type AppRouter struct {
	Router      *mux.Router
	DB          *sql.DB
	RestartChan chan struct{}
}
