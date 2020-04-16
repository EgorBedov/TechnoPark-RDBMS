package models

import "time"

type Thread struct {
	Id		int		`json:"id, omitempty"`
	Title 	string	`json:"title"`
	Author	string	`json:"author"`
	Forum	string	`json:"forum, omitempty"`
	Message string	`json:"message"`
	Votes	int		`json:"votes, omitempty"`
	Slug 	string	`json:"slug"`
	Created	time.Time	`json:"created, omitempty"`

	//AuthorId int32 	`json:"authorId, omitempty"`
	//ForumId	int32	`json:"forumId, omitempty"`
}
