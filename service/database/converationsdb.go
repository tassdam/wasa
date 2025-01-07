package database

import (
	"database/sql"
	"errors"
)

// GetUserConversations retrieves all conversations that a user participates in.
// It returns a slice of Conversation structs, each including its id, name, members, and lastMessage (if any).
func (db *appdbimpl) GetUserConversations(userId string) ([]Conversation, error) {
	var conversations []Conversation

	// Query all conversation IDs that the user is a member of
	rows, err := db.c.Query(`
		SELECT c.id, c.name
		FROM conversations c
		INNER JOIN conversation_members cm ON c.id = cm.conversationId
		WHERE cm.userId = ?
	`, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var conv Conversation
		if err := rows.Scan(&conv.Id, &conv.Name); err != nil {
			return nil, err
		}

		// Fetch members of the conversation
		members, err := db.getConversationMembers(conv.Id)
		if err != nil {
			return nil, err
		}
		conv.Members = members

		// Fetch the last message of the conversation, if any
		lastMessage, err := db.getLastMessage(conv.Id)
		if err != nil && !errors.Is(err, ErrMessageDoesNotExist) {
			// If there's another error, return it; if no last message, just ignore
			return nil, err
		}
		if lastMessage != nil {
			convWithLast := *lastMessage
			// Assign lastMessage to conversation, if found
			// For the summary, we only need one message object to represent lastMessage
			// Let's just store it as is:
			// Depending on your ConversationDetailsSummary schema,
			// you might need to store it differently or just assign it directly
			// Here we assume Conversation struct has a field 'LastMessage *Message' to hold it.
			// If not defined, update your Conversation struct accordingly.
			// For now, let's assume we can add a field in Conversation:
			//  LastMessage *Message `json:"lastMessage,omitempty"`
			// Make sure to update your Conversation struct in database.go accordingly.
			conv.LastMessage = &convWithLast
		}

		conversations = append(conversations, conv)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return conversations, nil
}

// GetConversationById retrieves full details of a specific conversation the user is part of.
// Returns a Conversation with id, name, members, and full messages list.
func (db *appdbimpl) GetConversationById(conversationId, userId string) (Conversation, error) {
	var conv Conversation

	// First, check that the user is a member of this conversation
	isMember, err := db.isUserMemberOfConversation(conversationId, userId)
	if err != nil {
		return conv, err
	}
	if !isMember {
		return conv, ErrConversationDoesNotExist // or a custom error indicating unauthorized access
	}

	// Fetch the conversation details
	if err := db.c.QueryRow(`SELECT id, name FROM conversations WHERE id = ?`, conversationId).Scan(&conv.Id, &conv.Name); err != nil {
		if err == sql.ErrNoRows {
			return conv, ErrConversationDoesNotExist
		}
		return conv, err
	}

	// Fetch members
	members, err := db.getConversationMembers(conv.Id)
	if err != nil {
		return conv, err
	}
	conv.Members = members

	// Fetch all messages in the conversation
	messages, err := db.getConversationMessages(conv.Id)
	if err != nil {
		return conv, err
	}
	// Assuming Conversation has a 'Messages []Message' field.
	conv.Messages = messages

	return conv, nil
}

// getConversationMembers is a helper function to retrieve the members of a conversation.
func (db *appdbimpl) getConversationMembers(conversationId string) ([]string, error) {
	var members []string
	rows, err := db.c.Query(`SELECT userId FROM conversation_members WHERE conversationId = ?`, conversationId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var userId string
		if err := rows.Scan(&userId); err != nil {
			return nil, err
		}
		members = append(members, userId)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return members, nil
}

// getLastMessage retrieves the last message (chronologically) of a conversation.
// Returns ErrMessageDoesNotExist if there are no messages.
func (db *appdbimpl) getLastMessage(conversationId string) (*Message, error) {
	var msg Message
	err := db.c.QueryRow(`
		SELECT id, senderId, content, timestamp, forwardedMessageId
		FROM messages
		WHERE conversationId=?
		ORDER BY timestamp DESC
		LIMIT 1
	`, conversationId).Scan(&msg.Id, &msg.SenderId, &msg.Content, &msg.Timestamp, &msg.ForwardedMessage)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrMessageDoesNotExist
		}
		return nil, err
	}
	return &msg, nil
}

// getConversationMessages retrieves all messages from a conversation in reverse chronological order.
// Returns an empty slice if no messages.
func (db *appdbimpl) getConversationMessages(conversationId string) ([]Message, error) {
	var msgs []Message
	rows, err := db.c.Query(`
		SELECT id, senderId, content, timestamp, forwardedMessageId
		FROM messages
		WHERE conversationId=?
		ORDER BY timestamp DESC
	`, conversationId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.Id, &msg.SenderId, &msg.Content, &msg.Timestamp, &msg.ForwardedMessage); err != nil {
			return nil, err
		}
		msgs = append(msgs, msg)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return msgs, nil
}

// isUserMemberOfConversation checks if a user is a member of a given conversation.
func (db *appdbimpl) isUserMemberOfConversation(conversationId, userId string) (bool, error) {
	var exists bool
	if err := db.c.QueryRow(`
		SELECT EXISTS(
		  SELECT 1 FROM conversation_members
		  WHERE conversationId = ? AND userId = ?
		)`, conversationId, userId).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}
