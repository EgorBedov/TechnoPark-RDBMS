package models

type Vote struct {
	Nickname	string		`json:"nickname"`
	Voice		int			`json:"voice"`
	ThreadSlugOrId	string	`json:"-"`
}
