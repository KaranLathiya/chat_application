package dataloader

import (
	"chat_application/api/dal"
	"chat_application/graph/model"
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
)

type ctxKeyType struct{ name string }

var CtxKey = ctxKeyType{"dataloaderctx"}

type Loaders struct {
	UserByID *UserLoader
}

func DataloaderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userloader := UserLoader{
			wait:     1 * time.Millisecond,
			maxBatch: 100,
			fetch: func(ids []string) ([]*model.Sender, []error) {
				var sqlQuery string
				if len(ids) == 1 {
					sqlQuery = "SELECT id, fullname FROM public.users WHERE id = ?"
				} else {
					sqlQuery = "SELECT id, fullname from public.users WHERE id IN (?)"
				}
				db := dal.GetDB()
				sqlQuery, arguments, err := sqlx.In(sqlQuery, ids)
				if err != nil {
					log.Println(err)
				}
				sqlQuery = sqlx.Rebind(sqlx.DOLLAR, sqlQuery)
				rows, err := dal.LogAndQuery(db, sqlQuery, arguments...)
				defer rows.Close()
				if err != nil {
					log.Println(err)
				}
				userById := map[string]*model.Sender{}

				for rows.Next() {
					user := model.Sender{}
					if err := rows.Scan(&user.ID, &user.Name); err != nil {

						return nil, []error{fmt.Errorf("internal server error")}
					}
					userById[user.ID] = &user
				}

				users := make([]*model.Sender, len(ids))
				for i, id := range ids {
					users[i] = userById[id]
					i++
				}

				return users, nil
			},
		}
		ctx := context.WithValue(r.Context(), CtxKey, &userloader)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
