package models

type User struct {
	NickName	string	`json:"nickname, omitempty "`
	FullName	string	`json:"fullname"`
	About		string	`json:"about, omitempty"`
	Email		string	`json:"email"`
}
