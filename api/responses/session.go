package responses

type SessionResponse struct {
	User  UserResponse `json:"user"`
	Token string       `json:"token"`
}
