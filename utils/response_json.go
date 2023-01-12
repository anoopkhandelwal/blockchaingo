package utils

import (
	"encoding/json"
	"log"
	"net/http"
)

func ToString(message string) string {
	bytes,err := json.Marshal(struct {
		Message	string	`json:"message"`
	}{Message:message})
	if err!=nil{
	  	  log.Println("Error Occured in response conversion")
	  	  return ""
	}
	return string(bytes[:])
}

func AddContentType(w http.ResponseWriter) http.ResponseWriter {
	w.Header().Add("Content-Type","application/json")
	return w
}
