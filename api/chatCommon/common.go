package chatCommon

import (
	"chat_application/api/dal"
	"fmt"
	"time"
)

func CurrentTimeConvertToCurrentFormattedTime() string {
	currentTime := time.Now()
	outputFormat := "2006-01-02 15:04:05-07:00"
	currentFormattedTime := currentTime.Format(outputFormat)
	return currentFormattedTime
}

func UserNameByID(userID string) (string, error) {
	db := dal.GetDB()
	var name string
	errIfNoRows := db.QueryRow("SELECT fullname FROM public.users WHERE id=$1", userID).Scan(&name)
	if errIfNoRows == nil {
		return name, nil
	}
	return "", fmt.Errorf("no user found with that id")
}
