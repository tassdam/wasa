package api

import "time"

// LoginRequest corresponds to the schema for user login requests.
type LoginRequest struct {
	Name string `json:"name"` // The name of the user to log in
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
	ID   string `json:"id"`   // Unique identifier of the user
	Name string `json:"name"` // Name of the user
}

// Message corresponds to the Message schema.
type Message struct {
	ID                 string    `json:"id"`                           // Unique identifier of the message
	SenderID           string    `json:"senderId"`                     // ID of the user who sent the message
	Content            string    `json:"content"`                      // Content of the message
	Timestamp          time.Time `json:"timestamp"`                    // When the message was sent
	ForwardedMessageID *string   `json:"forwardedMessageId,omitempty"` // ID of the original forwarded message (if any)
}

// ConversationDetailsSummary corresponds to a summary of a conversation.
type ConversationDetailsSummary struct {
	ID          string   `json:"id"`                    // Unique identifier of the conversation
	Name        string   `json:"name"`                  // Name of the conversation
	Members     []string `json:"members"`               // List of user IDs participating in the conversation
	LastMessage *Message `json:"lastMessage,omitempty"` // The last message in the conversation, if any
}

// ConversationDetails extends ConversationDetailsSummary by adding a messages array.
type ConversationDetails struct {
	ConversationDetailsSummary
	Messages []Message `json:"messages"` // List of messages in the conversation
}

// ConversationDetailsCollection represents a list of conversations.
type ConversationDetailsCollection struct {
	Conversations []ConversationDetailsSummary `json:"conversations"` // List of conversation summaries
}

// SendMessageRequest is used to send a message.
type SendMessageRequest struct {
	Content            string  `json:"content"`                      // The content of the message
	ForwardedMessageID *string `json:"forwardedMessageId,omitempty"` // ID of the forwarded message, if any
}

// ForwardMessageRequest is used to forward a message.
type ForwardMessageRequest struct {
	DestinationConversationID string `json:"destinationConversationId"` // The conversation ID to which the message is forwarded
}

// AddCommentRequest is used to add a comment to a message.
type AddCommentRequest struct {
	Content string `json:"content"` // The content of the comment
}

// Comment corresponds to a comment on a message.
type Comment struct {
	ID        string    `json:"id"`        // Unique identifier of the comment
	AuthorID  string    `json:"authorId"`  // ID of the user who wrote the comment
	Content   string    `json:"content"`   // Content of the comment
	Timestamp time.Time `json:"timestamp"` // When the comment was made
}

// AddGroupMemberRequest is used to add a user to a group.
type AddGroupMemberRequest struct {
	UserID string `json:"userId"` // ID of the user to add
}

// GroupMember represents a member of a group.
type GroupMember struct {
	UserID   string    `json:"userId"`   // ID of the user
	JoinedAt time.Time `json:"joinedAt"` // When the user joined the group
}

// UpdateGroupRequest is used to update a group's information.
type UpdateGroupRequest struct {
	Name string `json:"name"` // New name of the group
}

// Group corresponds to a group schema.
type Group struct {
	ID   string `json:"id"`   // Unique identifier of the group
	Name string `json:"name"` // Name of the group
}
