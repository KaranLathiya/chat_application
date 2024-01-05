package errors

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/lib/pq"
)

type Message struct {
	Code    int    `json:"code"  validate:"required"`
	Message string `json:"message"  validate:"required"`
}

func CheckErr(err error) {
	if err != nil {
		panic(err)
	}
}

// func ErrorShow(code int, message string) []byte {
// 	errMessage.Code = code
// 	errMessage.Message = message
// 	user_data, _ := json.MarshalIndent(errMessage, "", "  ")
// 	return user_data
// }

func MessageShow(code int, message string, w http.ResponseWriter) {
	var Message Message
	Message.Code = code
	Message.Message = message
	user_data, _ := json.MarshalIndent(Message, "", "  ")
	w.WriteHeader(code)
	w.Write(user_data)
}

func DatabaseErrorShow(err error) (string, int) {
	if dbErr, ok := err.(*pq.Error); ok { // For PostgreSQL database driver (pq)
		// Access PostgreSQL-specific error fields
		// errCode,_ :=  strconv.Atoi(dbErr.Code)
		errCode := dbErr.Code
		errMessage := errCode.Name()
		errDetail := dbErr.Detail
		// Handle the PostgreSQL-specific error
		fmt.Println(errCode)
		fmt.Println(errDetail)
		fmt.Println(errMessage)
		switch errCode {
		case "23502":
			// not-null constraint violation
			return "Some required data was left out", 400

		case "23503":
			// foreign key violation
			return "This record can't be changed because another record refers to it", 409

		case "23505":
			// unique constraint violation
			return "This record contains duplicated data that conflicts with what is already in the database", 409

		case "23514":
			// check constraint violation
			return "This record contains inconsistent or out-of-range data", 400

		}
	}
	return err.Error(), 401
}
