package user

import (
	"chat_application/api/auth"
	"chat_application/api/customError"
	"chat_application/api/dal"
	"chat_application/graph/model"
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

var ConversationNotificationMap = make(map[string]chan model.ConversationNotification)

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
	// fmt.Println(query)
	rows, err := db.Query(query, filterArgsList...)
	if err != nil {
		return nil, err
	}
	var users []*model.User
	for rows.Next() {
		var user model.User
		err = rows.Scan(&user.ID, &user.FullName, &user.Email, &user.IPAddress, &user.Gender)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}
	return users, nil
}

func RecentConversationList(ctx context.Context, limit *int, offset *int) ([]*model.ConversationList, error) {
	db := dal.GetDB()
	userID := ctx.Value(auth.UserCtxKey).(string)
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
		LIMIT $2 OFFSET $3;`, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	var conversationList []*model.ConversationList
	for rows.Next() {
		var conversation model.ConversationList
		err = rows.Scan(&conversation.LastMessageTime, &conversation.ConversationID, &conversation.IsItGroup)
		if err != nil {
			return nil, err
		}
		conversationList = append(conversationList, &conversation)
	}
	return conversationList, nil
}

func ConversationNotification(ctx context.Context) (<-chan model.ConversationNotification, error) {
	userID := ctx.Value(auth.UserCtxKey).(string)
	conversationEvent, ok := ConversationNotificationMap[userID]
	fmt.Println(conversationEvent)
	if ok {
		close(conversationEvent)
	}
	go func() {
		_, ok := <-conversationEvent
		if !ok {
			return
		}
		<-ctx.Done()
		defer clearSubscriptionVariables(userID)
		// select {
		// case _, ok := <-conversationEvent:
		// 	if !ok {
		// 		ctx.Done()
		// 		return
		// 	}
		// case <-ctx.Done():
		// 	defer clearSubscriptionVariables(userID)
		// 	return
		// }
	}()
	fmt.Println("new chan")
	conversationEvent = make(chan model.ConversationNotification, 1)
	ConversationNotificationMap[userID] = conversationEvent
	// fmt.Println("after allocating variable")
	fmt.Println(conversationEvent)
	// printAllocatedMemory()
	// runtime.KeepAlive(senderAndReceiverMap) // Keeps a reference to m so that the map isn’t collected
	return conversationEvent, nil
}

func UserDetailsByID(ctx context.Context, userID string) (*model.User, error) {
	db := dal.GetDB()
	var user model.User
	errIfNoRows := db.QueryRow("SELECT fullname, email, ip_address, gender FROM public.users WHERE id=$1", userID).Scan(&user.FullName, &user.Email, &user.IPAddress, &user.Gender)
	if errIfNoRows != nil {
		if errIfNoRows.Error() == "sql: no rows in result set" {
			return nil, fmt.Errorf("no user found with that Id")
		}
		databaseErrorMessage := customError.DatabaseErrorShow(errIfNoRows)
		return nil, fmt.Errorf(databaseErrorMessage)
	}
	user.ID = userID
	return &user, nil
}

func UserNameByID(ctx context.Context, userID string) (string, error) {
	db := dal.GetDB()
	var name string
	errIfNoRows := db.QueryRow("SELECT fullname FROM public.users WHERE id=$1", userID).Scan(&name)
	if errIfNoRows != nil {
		if errIfNoRows.Error() == "sql: no rows in result set" {
			return "", fmt.Errorf("no user found with that Id")
		}
		databaseErrorMessage := customError.DatabaseErrorShow(errIfNoRows)
		return "", fmt.Errorf(databaseErrorMessage)
	}
	return name, nil
}

func clearSubscriptionVariables(id string) {
	fmt.Println("clearing")
	delete(ConversationNotificationMap, id)
}
