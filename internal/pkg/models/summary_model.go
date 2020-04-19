package models

type Summary struct {
	Users	int		`json:"user"`
	Forums	int		`json:"forum"`
	Threads	int		`json:"thread"`
	Posts	int		`json:"post"`
}