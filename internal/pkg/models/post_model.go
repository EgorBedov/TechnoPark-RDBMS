package models

import "time"

type Post struct {
	Id       	int       `json:"id, omitempty"`
	Parent   	int       `json:"parent"`
	Author   	string    `json:"author"`
	Message  	string    `json:"message"`
	IsEdited 	bool      `json:"isEdited, omitempty"`
	Forum    	string    `json:"forum, omitempty"`
	ThreadId 	int       `json:"thread, omitempty"`
	Created  	time.Time `json:"created, omitempty"`
}

type PostInfo struct {
	Pst 		*Post		`json:"post"`
	Author		*User		`json:"author,omitempty"`
	Thrd		*Thread		`json:"thread,omitempty"`
	Frm 		*Forum		`json:"forum,omitempty"`
}
