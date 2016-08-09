package models

import (
	"encoding/json"
	"log"
)

type Email struct {
	Ip   string
	Data string
}

func MakeEmail(ip string, data []byte) Email {
	jsonData, err := json.Marshal(data)

	if err != nil {
		log.Println(err)
	}
	c := Email{
		Ip:   ip,
		Data: string(jsonData),
	}
	return c
}
