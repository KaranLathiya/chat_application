package auth

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
)

var UserCtxKey = &contextKey{"user"}

type contextKey struct {
	name string
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("in middleware")
		header := r.Header.Get("Authorization")
		fmt.Println(header)
		// Allow unauthenticated users in
		if header == "" {
			fmt.Println("no header")
			ctx := context.WithValue(r.Context(), UserCtxKey, 0)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		AuthorizationKey,_ := strconv.Atoi(header)
		// // put it in context
			
		ctx := context.WithValue(r.Context(), UserCtxKey, AuthorizationKey)

		// and call the next with our new context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
