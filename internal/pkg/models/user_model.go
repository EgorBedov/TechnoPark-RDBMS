package models

type User struct {
	NickName	string	`json:"nickname, omitempty "`
	FullName	string	`json:"fullname"`
	About		string	`json:"about, omitempty"`
	Email		string	`json:"email"`
}

//func NewUser() Model {
//	return &User{}
//}
//
//func (u *User) Marshal() ([]byte, error) {
//	return json.Marshal(User{
//		NickName: u.NickName,
//		FullName: u.FullName,
//		About: u.About,
//		Email: u.Email,
//	})
//}
