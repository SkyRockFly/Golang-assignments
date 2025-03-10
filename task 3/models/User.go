package models

import "encoding/json"

//easyjson:json
type User struct {
	Browsers []string `json:"browsers"`
	Email    string   `json:"email"`
	Name     string   `json:"name"`
}

//easyjson:json
type LazyUser struct {
	Browsers json.RawMessage `json:"browsers"`
}

//Job      string   `json:"job"`
//Phone    string   `json:"phone"`
//Company  string   `json:"company"`
//Country  string   `json:"country"`
