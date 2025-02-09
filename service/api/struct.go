package api

import "time"

type LoginRequest struct {
	Name  string `json:"name"`
	Photo string `json:"photo"`
}

type LoginResponse struct {
	Identifier string `json:"identifier"`
}

type UpdateUserRequest struct {
	Name string `json:"name"`
}

type UpdateGroupRequest struct {
	Name string `json:"groupName"`
}

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Photo []byte `json:"photo"`
}

type Message struct {
	ID        string    `json:"id"`
	SenderID  string    `json:"senderId"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}
