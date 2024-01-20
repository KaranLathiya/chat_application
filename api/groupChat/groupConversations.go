package groupChat

import (
	"chat_application/api/auth"
	"chat_application/api/chatCommon"
	"chat_application/api/dal"
	"chat_application/graph/model"
	"context"
	"database/sql"
	"fmt"

	"github.com/markbates/going/randx"
)

func init() {
	groupConversationPublishedChannelMap = map[string]chan *model.GroupConversation{}
}

var groupAndMemberMap = make(map[string]map[string]string)
var groupConversationPublishedChannelMap map[string]chan *model.GroupConversation

func CreateGroupConversation(ctx context.Context, input model.NewGroupConversation) (*model.GroupConversation, error) {
	var GroupConversation model.GroupConversation
	db := dal.GetDB()
	senderID := ctx.Value(auth.UserCtxKey).(string)
	var removedFromGroup bool
	errIfNoRows := db.QueryRow(
		"SELECT is_removed FROM public.group_members WHERE member_id=$1 AND group_id=$2;", senderID, input.GroupID).Scan(&removedFromGroup)
	if errIfNoRows != nil {
		return nil, fmt.Errorf("user is not member of group")
	}
	if removedFromGroup {
		return nil, fmt.Errorf("user is no more member of group")
	}
	currentFormattedTime := chatCommon.CurrentTimeConvertToCurrentFormattedTime()
	errIfNoRows = db.QueryRow(
		"INSERT INTO public.group_conversations( group_id, sender_id, content, created_at) VALUES ( $1, $2, $3, $4)  RETURNING id, created_at;", input.GroupID, senderID, input.Content, currentFormattedTime).Scan(&GroupConversation.ID, &GroupConversation.CreatedAt)
	if errIfNoRows == nil {
		GroupConversation.SenderID = senderID
		GroupConversation.GroupID = input.GroupID
		GroupConversation.Content = input.Content
		go func() {
			for id, _ := range groupAndMemberMap {
				fmt.Println("sub running")
				if groupAndMemberMap[id]["groupID"] == input.GroupID {
					groupConversationPublishedChannelMap[id] <- &GroupConversation
				}
			}
		}()
		return &GroupConversation, nil
	}
	return nil, errIfNoRows
}

func GroupConversationRecords(ctx context.Context, limit *int, offset *int, groupID string) ([]*model.GroupConversation, error) {
	db := dal.GetDB()
	UserID := ctx.Value(auth.UserCtxKey).(string)
	var removedFromGroup bool
	var removedAt *string
	errIfNoRows := db.QueryRow(
		"SELECT is_removed,removed_at FROM public.group_members WHERE member_id=$1 AND group_id=$2;", UserID, groupID).Scan(&removedFromGroup, &removedAt)
	if errIfNoRows != nil {
		return nil, fmt.Errorf("invalid groupid or memberid")
	}
	var rows *sql.Rows
	var err error
	if removedFromGroup {
		rows, err = db.Query("SELECT id, sender_id, content, created_at FROM public.group_conversations WHERE group_id = $1 AND created_at <= $2 ORDER BY created_at DESC LIMIT $3 OFFSET $4 ", groupID, *removedAt, limit, offset)
	} else {
		rows, err = db.Query("SELECT id, sender_id, content, created_at FROM public.group_conversations WHERE group_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3 ", groupID, limit, offset)
	}
	if err != nil {
		return nil, err
	}
	var GroupConversations []*model.GroupConversation
	for rows.Next() {
		var GroupConversation model.GroupConversation
		GroupConversation.GroupID = groupID
		err = rows.Scan(&GroupConversation.ID, &GroupConversation.SenderID, &GroupConversation.Content, &GroupConversation.CreatedAt)
		if err != nil {
			return nil, err
		}
		GroupConversations = append(GroupConversations, &GroupConversation)
	}
	return GroupConversations, nil
}

func DeleteGroupConversation(ctx context.Context, input model.DeleteGroupConversationInput) (bool, error) {
	senderID := ctx.Value(auth.UserCtxKey).(string)
	db := dal.GetDB()
	var removedFromGroup bool
	errIfNoRows := db.QueryRow(
		"SELECT is_removed FROM public.group_members WHERE member_id=$1 AND group_id=$2;", senderID, input.GroupID).Scan(&removedFromGroup)
	if errIfNoRows != nil {
		return false, fmt.Errorf("invalid groupid or memberid")
	}
	if removedFromGroup {
		return false, fmt.Errorf("user is no more member of group")
	}
	result, err := db.Exec("DELETE FROM public.group_conversations WHERE sender_id=$1 AND id=$2;", senderID, input.MessageID)
	if err != nil {
		return false, err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return false, fmt.Errorf("invalid senderid or messageid")
	}
	return true, nil
}

func GroupConversationPublished(ctx context.Context, input model.GroupConversationPublishedInput) (<-chan *model.GroupConversation, error) {
	db := dal.GetDB()
	var removedFromGroup bool
	errIfNoRows := db.QueryRow(
		"SELECT is_removed FROM public.group_members WHERE member_id=$1 AND group_id=$2;", input.MemberID, input.GroupID).Scan(&removedFromGroup)
	if errIfNoRows != nil {
		return nil, fmt.Errorf("invalid groupid or memberid")
	}
	if removedFromGroup {
		return nil, fmt.Errorf("user is no more member of group")
	}
	id := randx.String(8)
	// fmt.Println(id)
	fmt.Println("GroupConversationPublished running")
	groupConversationEvent := make(chan *model.GroupConversation, 1)
	go func() {
		<-ctx.Done()
		defer clearSubscriptionVariablesOfGroupConversation(id)
	}()
	groupAndMemberMap[id] = map[string]string{"groupID": input.GroupID}
	groupConversationPublishedChannelMap[id] = groupConversationEvent
	return groupConversationEvent, nil
}

func clearSubscriptionVariablesOfGroupConversation(id string) {
	delete(groupAndMemberMap, id)
	delete(groupConversationPublishedChannelMap, id)
}
