package models

type Forum struct {
	Title 	*string	`json:"title"`
	Usr 	*string	`json:"user"`
	Slug 	string	`json:"slug"`
	Posts 	*int64	`json:"posts, omitempty"`
	Threads *int32	`json:"threads, omitempty"`
}
