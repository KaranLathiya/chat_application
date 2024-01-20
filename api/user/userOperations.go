package user

import (
	"chat_application/api/auth"
	"chat_application/api/dal"
	"chat_application/graph/model"
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

func UserList(ctx context.Context, input *model.UserListInput) ([]*model.User, error) {
	db := dal.GetDB()
	offset := (*input.Page - 1) * *input.Limit
	var where, orderBy []string
	var whereKeyword string
	var filterArgsList []interface{}

	if input.Name != nil && *input.Name != "" {
		where = append(where, "fullname ILIKE '%' || ? || '%'")
		filterArgsList = append(filterArgsList, *input.Name)
		orderBy = append(orderBy, "POSITION (LOWER('"+*input.Name+"') IN LOWER(fullname)) ASC")
	}
	if input.Email != nil && *input.Email != "" {
		where = append(where, "email ILIKE '%' || ? || '%'")
		filterArgsList = append(filterArgsList, *input.Email)
		orderBy = append(orderBy, "POSITION (LOWER('"+*input.Email+"') IN LOWER(email)) ASC")
	}
	if len(where) > 0 {
		whereKeyword = "WHERE"
	} else {
		orderBy = append(orderBy, "fullname ASC")
	}
	// fmt.Println(wh)
	query := fmt.Sprintf("SELECT id, fullname, email, ip_address, gender FROM public.users %s %v ORDER BY %v LIMIT %d OFFSET %d", whereKeyword, strings.Join(where, " OR "), strings.Join(orderBy, " , "), *input.Limit, offset)
	query = sqlx.Rebind(sqlx.DOLLAR, query)
	fmt.Println(query)
	rows, err := db.Query(query, filterArgsList...)
	if err != nil {
		return nil, err
	}
	var Users []*model.User
	for rows.Next() {
		var User model.User
		err = rows.Scan(&User.ID, &User.Fullname, &User.Email, &User.IPAddress, &User.Gender)
		if err != nil {
			return nil, err
		}
		Users = append(Users, &User)
	}
	return Users, nil
}

func RecentConversationList(ctx context.Context, limit *int, offset *int) ([]*model.ConversationList, error) {
	db := dal.GetDB()
	UserID := ctx.Value(auth.UserCtxKey).(string)
	rows, err := db.Query(
		`
		SELECT
			last_message_time,
			conversation_id,
			is_it_group
		FROM
			(
			SELECT
				CASE
					WHEN is_removed = true THEN removed_at
					ELSE MAX(created_at)
				END AS last_message_time,
				gc.group_id AS conversation_id,
				'true' AS is_it_group
			FROM
				public.group_conversations gc
			INNER JOIN public.group_members gm ON
				gc.group_id = gm.group_id
			WHERE
				gc.group_id IN (
				SELECT
					group_id
				FROM
					public.group_members
				WHERE
					member_id = $1
		    )
				AND gm.member_id = $1
			GROUP BY
				(gc.group_id,
				gm.is_removed,
				gm.removed_at)
		UNION
			SELECT
				MAX(created_at) AS last_message_time,
				CASE
					WHEN sender_id = $1 THEN receiver_id
					ELSE sender_id
				END AS recent_conversation_id,
				'false' AS is_it_group
			FROM
				public.personal_conversations
			WHERE
				sender_id = $1
				OR receiver_id = $1
			GROUP BY
				recent_conversation_id
		)
		ORDER BY
			last_message_time DESC
		LIMIT $2 OFFSET $3;`, UserID, limit, offset)
	if err != nil {
		return nil, err
	}
	var ConversationList []*model.ConversationList
	for rows.Next() {
		var Conversation model.ConversationList
		err = rows.Scan(&Conversation.LastMessageTime, &Conversation.ConversationID, &Conversation.IsItGroup)
		if err != nil {
			return nil, err
		}
		ConversationList = append(ConversationList, &Conversation)
	}
	return ConversationList, nil
}

func UserDetailsByID(ctx context.Context, userID string) (*model.User, error) {
	db := dal.GetDB()
	id := ctx.Value(auth.UserCtxKey).(string)
	var User model.User
	errIfNoRows := db.QueryRow("SELECT fullname, email, ip_address, gender FROM public.users WHERE id=$1", id).Scan(&User.Fullname, &User.Email, &User.IPAddress, &User.Gender)
	if errIfNoRows == nil {
		User.ID = id
		return &User, nil
	}
	return &User, errIfNoRows
}

func UserNameByID(ctx context.Context, userID string) (string, error) {
	db := dal.GetDB()
	var name string
	errIfNoRows := db.QueryRow("SELECT fullname FROM public.users WHERE id=$1", userID).Scan(&name)
	if errIfNoRows == nil {
		return name, nil
	}
	return "", fmt.Errorf("no user found with that id")
}