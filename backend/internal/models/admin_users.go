package models

type AdminUserPage struct {
	Users  []UserDTO `json:"users"`
	Total  int64     `json:"total"`
	Limit  int       `json:"limit"`
	Offset int       `json:"offset"`
}
