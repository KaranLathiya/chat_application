package graph

import (
	"chat_application/api/auth"
	"chat_application/api/customError"
	"chat_application/api/dal"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	_ "github.com/99designs/gqlgen/graphql/introspection"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	// db *sql.DB
}

func NewRootResolvers(db *sql.DB) Config {
	c := Config{
		Resolvers: &Resolver{
			// db: db
		},
	}
	// Complexity
	// Schema Directive
	c.Directives.IsAuthenticated = func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
		authorizationKey := ctx.Value(auth.UserCtxKey).(string)
		// fmt.Println("Authorization key:" + authorizationKey)
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
	errIfNoRows := db.QueryRow("SELECT id from public.users where id=$1", id).Scan(&id)
	if errIfNoRows != nil {
		if errIfNoRows.Error() == "sql: no rows in result set" {
			return false, "invalid authorization key"
		}
		errMessage := customError.DatabaseErrorShow(errIfNoRows)
		return false, errMessage
	}
	return true,""
}
