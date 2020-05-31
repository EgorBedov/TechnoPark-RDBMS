package models

type Message struct {
	Error 		error		`json:"-"`
	Message 	string		`json:"message"`
	Status		int			`json:"-"`
}

func CreateError(err error, msg string, code int) Message {
	return Message{
		Error:   err,
		Message: msg,
		Status:  code,
	}
}

func CreateSuccess(code int) Message {
	return Message{
		Error:   nil,
		Message: "",
		Status:  code,
	}
}
