package _115

type GetFileIdResponse struct {
	State     bool        `json:"state"`
	Error     string      `json:"error"`
	Errno     int         `json:"errno"`
	Id        interface{} `json:"id"`
	IsPrivate interface{} `json:"is_private"`
}
