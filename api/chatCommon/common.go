package chatCommon

import (
	"time"
)

// type Sender struct {
// 	Name string `json:"name"`
// 	ID   string `json:"id"`
// }
// type GroupConversationWithSender struct {
// 	ID        string `json:"id"`
// 	GroupID   string `json:"groupId"`
// 	Content   string `json:"content"`
// 	CreatedAt string `json:"createdAt"`
// 	SenderID  string `json:"-"`
// }

func CurrentTimeConvertToCurrentFormattedTime() string {
	// fmt.Println(time.Now().UTC())
	// fmt.Println(time.Now().Local().UTC())
	currentTime := time.Now().UTC()
	outputFormat := "2006-01-02 15:04:05-07:00"
	currentFormattedTime := currentTime.Format(outputFormat)
	// fmt.Println(currentFormattedTime)
	return currentFormattedTime
}
