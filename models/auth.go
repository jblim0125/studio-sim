package models

// AuthReqBody POST angora/auth, angora/v2/auth request body
type AuthReqBody struct {
	Password  string `json:"password"`
	Encrypted bool   `json:"encrypted"`
	GroupID   string `json:"group_id"`
	ID        string `json:"id"`
}

// AuthResBody angora/auth, angora/v2/auth response body
type AuthResBody struct {
	Status string `json:"status"`
	Token  string `json:"token"`
}
