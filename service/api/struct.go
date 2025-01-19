package api

import "time"

type LoginRequest struct {
	Name  string `json:"name"`
	Photo string `json:"photo"`
}

// LoginResponse corresponds to the schema for user login responses.
type LoginResponse struct {
	Identifier string `json:"identifier"` // The identifier of the logged-in user
}

// UpdateUserRequest is used for updating the authenticated user's information.
type UpdateUserRequest struct {
	Name string `json:"name"` // New username
}

// User corresponds to the User schema.
type User struct {
	ID    string `json:"id"`   // Unique identifier of the user
	Name  string `json:"name"` // Name of the user
	Photo []byte `json:"photo"`
}

// Message corresponds to the Message schema.
type Message struct {
	ID                 string    `json:"id"`                           // Unique identifier of the message
	SenderID           string    `json:"senderId"`                     // ID of the user who sent the message
	Content            string    `json:"content"`                      // Content of the message
	Timestamp          time.Time `json:"timestamp"`                    // When the message was sent
	ForwardedMessageID *string   `json:"forwardedMessageId,omitempty"` // ID of the original forwarded message (if any)
}
