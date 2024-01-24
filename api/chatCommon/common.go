package chatCommon

import (
	"time"
)

func CurrentTimeConvertToCurrentFormattedTime() string {
	// fmt.Println(time.Now().UTC())
	// fmt.Println(time.Now().Local().UTC())
	currentTime := time.Now()
	outputFormat := "2006-01-02 15:04:05-07:00"
	currentFormattedTime := currentTime.Format(outputFormat)
	// fmt.Println(currentFormattedTime)
	return currentFormattedTime
}
