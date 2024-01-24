package auth

import (
	"context"
	"fmt"
	"net/http"
)

var UserCtxKey = &contextKey{"user"}

type contextKey struct {
	name string
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("in middleware")
		handleCORS(w,r)
		authorizationKey := r.Header.Get("Authorization")
		fmt.Println("Authorization key:"+authorizationKey)
		
		// Allow unauthenticated users in
		if authorizationKey == "" {
			ctx := context.WithValue(r.Context(), UserCtxKey, "")
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		ctx := context.WithValue(r.Context(), UserCtxKey, authorizationKey)

		// and call the next with our new context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func handleCORS(w http.ResponseWriter,r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Content-Type", "json/application")
	// Handle preflight requests
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
}
