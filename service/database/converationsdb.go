package database

import (
	"database/sql"
	"fmt"
)

// GetMyConversations returns the list of conversations in which userID is a member.
func (db *appdbimpl) GetMyConversations(userID string) ([]Conversation, error) {
	var convs []Conversation

	// Query all conversation IDs for which userID is in conversation_members
	rows, err := db.c.Query(`
        SELECT c.id, c.name
        FROM conversations c
        JOIN conversation_members cm ON c.id = cm.conversationId
        WHERE cm.userId = ?
        -- OPTIONAL: you can ORDER BY last message timestamp if desired
        -- ORDER BY (SELECT MAX(timestamp) FROM messages WHERE conversationId = c.id) DESC
    `, userID)
	if err != nil {
		return nil, fmt.Errorf("GetMyConversations query error: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var conv Conversation
		if err := rows.Scan(&conv.Id, &conv.Name); err != nil {
			return nil, err
		}

		// Fill the Members slice
		members, memErr := db.getConversationMembers(conv.Id)
		if memErr != nil {
			return nil, memErr
		}
		conv.Members = members

		// Optionally fetch last message
		lastMsg, lmErr := db.getLastMessage(conv.Id)
		if lmErr != nil && lmErr != sql.ErrNoRows {
			// If it's something other than "no messages" error, return it
			return nil, lmErr
		}
		if lastMsg != nil {
			conv.LastMessage = lastMsg
		}

		// We'll ignore conv.Messages here, so it remains empty unless needed
		convs = append(convs, conv)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return convs, nil
}

// getConversationMembers returns the user IDs of all members in the conversation.
func (db *appdbimpl) getConversationMembers(conversationID string) ([]string, error) {
	var members []string

	rows, err := db.c.Query(`
        SELECT userId
        FROM conversation_members
        WHERE conversationId = ?
    `, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err != nil {
			return nil, err
		}
		members = append(members, uid)
	}
	// check rows.Err()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return members, nil
}

// getLastMessage returns the most recent message for a conversation, or nil if none exist.
func (db *appdbimpl) getLastMessage(conversationID string) (*Message, error) {
	var msg Message
	err := db.c.QueryRow(`
        SELECT id, conversationId, senderId, content, timestamp, forwardedMessageId
        FROM messages
        WHERE conversationId = ?
        ORDER BY timestamp DESC
        LIMIT 1
    `, conversationID).Scan(
		&msg.Id,
		&msg.ConversationId,
		&msg.SenderId,
		&msg.Content,
		&msg.Timestamp,
		&msg.ForwardedMessage,
	)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}
