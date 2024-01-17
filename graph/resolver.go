package graph

import (
	"chat_application/api/auth"
	"chat_application/api/dal"
	"context"
	"errors"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	_ "github.com/99designs/gqlgen/graphql/introspection"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct{}

func NewRootResolvers() Config {
	c := Config{
		Resolvers: &Resolver{},
	}
	// Complexity
	// Schema Directive
	c.Directives.IsAuthenticated = func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
		authorizationKey := ctx.Value(auth.UserCtxKey).(string)
		fmt.Println("Authorization key:" + authorizationKey)
		if authorizationKey != "" {
			fmt.Println("with autho")
			ok, errorMessage := validateUserByAuthorizationKey(authorizationKey)
			if ok {
				return next(ctx)
			} else {
				return nil, errors.New(errorMessage)
			}
		} else {
			fmt.Println("no autho")
			return nil, errors.New("no authorization key")
		}
	}
	return c
}
func validateUserByAuthorizationKey(id string) (bool, string) {
	db := dal.GetDB()
	rows, err := db.Query("select id from public.users where id=$1", id)
	if err != nil {
		return false, "internal server error"
	}
	i := 0
	for rows.Next() {
		i += 1
	}
	defer rows.Close()
	if i == 0 {
		return false, "invalid authorization key"
	}
	return true, ""
}
