package personalChat

import (
	"chat_application/api/auth"
	"chat_application/api/chatCommon"
	"chat_application/api/dal"
	"chat_application/graph/model"
	"context"
	"fmt"

	"github.com/markbates/going/randx"
)

// func init() {
// 	personalConversationPublishedChannelMap = map[string]chan *model.PersonalConversation{}
// }

var personalConversationPublishedChannelMap = make(map[string]chan *model.PersonalConversation)
var senderAndReceiverMap = make(map[string]map[string]string)

func PersonalConversationRecords(ctx context.Context, limit *int, offset *int, receiverID string) ([]*model.PersonalConversation, error) {
	db := dal.GetDB()
	userID := ctx.Value(auth.UserCtxKey).(string)
	rows, err := db.Query(
		"SELECT sender_id, receiver_id, content, created_at, id FROM public.personal_conversations WHERE (sender_id = $1 AND receiver_id = $2) OR (sender_id = $2 AND receiver_id = $1) ORDER BY created_at DESC LIMIT $3 OFFSET $4 ", userID, receiverID, limit, offset)
	if err != nil {
		return nil, err
	}
	var personalConversations []*model.PersonalConversation
	for rows.Next() {
		var personalConversation model.PersonalConversation
		err = rows.Scan(&personalConversation.SenderID, &personalConversation.ReceiverID, &personalConversation.Content, &personalConversation.CreatedAt, &personalConversation.ID)
		if err != nil {
			return nil, err
		}
		personalConversations = append(personalConversations, &personalConversation)
	}
	return personalConversations, nil
}

func CreatePersonalConversation(ctx context.Context, input model.NewPersonalConversation) (*model.PersonalConversation, error) {
	var personalConversation model.PersonalConversation
	db := dal.GetDB()
	senderID := ctx.Value(auth.UserCtxKey).(string)
	currentFormattedTime := chatCommon.CurrentTimeConvertToCurrentFormattedTime()
	errIfNoRows := db.QueryRow(
		"INSERT INTO public.personal_conversations( sender_id, receiver_id, content, created_at) VALUES ( $1, $2, $3, $4)  RETURNING id, created_at;", senderID, input.ReceiverID, input.Content, currentFormattedTime).Scan(&personalConversation.ID, &personalConversation.CreatedAt)
	if errIfNoRows == nil {
		personalConversation.SenderID = senderID
		personalConversation.ReceiverID = input.ReceiverID
		personalConversation.Content = input.Content
		go func() {
			for id, _ := range senderAndReceiverMap {
				fmt.Println("sub running")
				if (senderAndReceiverMap[id]["senderID"] == senderID && senderAndReceiverMap[id]["receiverID"] == input.ReceiverID) || (senderAndReceiverMap[id]["senderID"] == input.ReceiverID && senderAndReceiverMap[id]["receiverID"] == senderID) {
					fmt.Println("sender and receiver varified")
					personalConversationPublishedChannelMap[id] <- &personalConversation
				}
			}
		}()
		return &personalConversation, nil
	}
	return nil, errIfNoRows
}

func DeletePersonalConversation(ctx context.Context, messageID string) (bool, error) {
	senderID := ctx.Value(auth.UserCtxKey).(string)
	db := dal.GetDB()
	result, err := db.Exec("DELETE FROM public.personal_conversations WHERE sender_id=$1 AND id=$2;", senderID, messageID)
	if err != nil {
		return false, err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return false, fmt.Errorf("wrong data")
	}
	return true, nil
}

func PersonalConversationNotification(ctx context.Context, input model.PersonalConversationNotificationInput) (<-chan *model.PersonalConversation, error) {
	id := randx.String(8)
	// fmt.Println(id)
	// printAllocatedMemory()
	personalConversationEvent := make(chan *model.PersonalConversation, 1)
	go func() {
		<-ctx.Done()
		defer clearSubscriptionVariablesOfPersonalConversation(id)
	}()
	senderAndReceiverMap[id] = map[string]string{"senderID": input.SenderID, "receiverID": input.ReceiverID}
	personalConversationPublishedChannelMap[id] = personalConversationEvent
	// fmt.Println("after allocating variable"p

	// printAllocatedMemory()
	// runtime.KeepAlive(senderAndReceiverMap) // Keeps a reference to m so that the map isn’t collected
	return personalConversationEvent, nil
}

func clearSubscriptionVariablesOfPersonalConversation(id string) {
	delete(senderAndReceiverMap, id)
	delete(personalConversationPublishedChannelMap, id)
}
