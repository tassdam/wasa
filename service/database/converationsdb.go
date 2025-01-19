package database

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"time"
)

func (db *appdbimpl) GetDirectConversation(senderID, recipientID string) (string, error) {
	var conversationID string

	// Query for existing direct conversation
	err := db.c.QueryRow(`
		SELECT id
		FROM conversations
		WHERE type = 'direct'
		  AND id IN (
		      SELECT conversationId FROM conversation_members WHERE userId = ?
		      INTERSECT
		      SELECT conversationId FROM conversation_members WHERE userId = ?
		  )
	`, senderID, recipientID).Scan(&conversationID)

	if err == sql.ErrNoRows {
		return "", nil // No conversation exists
	}
	if err != nil {
		return "", fmt.Errorf("error checking conversation: %w", err)
	}

	return conversationID, nil // Return the existing conversation ID
}

func (db *appdbimpl) CreateDirectConversation(conversationID, senderID, recipientID string) error {
	// Insert a new conversation

	_, err := db.c.Exec(`
		INSERT INTO conversations (id, name, type, created_at, conversationPhoto)
		VALUES (?, '', 'direct', ?, '')
	`, conversationID, time.Now().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("error creating new conversation: %w", err)
	}

	// Add both users to the conversation
	_, err = db.c.Exec(`
		INSERT INTO conversation_members (conversationId, userId)
		VALUES (?, ?), (?, ?)
	`, conversationID, senderID,
		conversationID, recipientID)
	if err != nil {
		return fmt.Errorf("error adding members to conversation_members: %w", err)
	}

	return nil
}

func (db *appdbimpl) SaveMessage(
	conversationID, senderID, messageID, content string,
	forwardedMessageID *string, attachment []byte,
) (Message, error) {
	// Check if the conversation exists
	var conversationExists bool
	err := db.c.QueryRow(`SELECT EXISTS(SELECT 1 FROM conversations WHERE id = ?)`, conversationID).Scan(&conversationExists)
	if err != nil {
		return Message{}, fmt.Errorf("error checking conversation existence: %w", err)
	}
	if !conversationExists {
		return Message{}, ErrConversationDoesNotExist
	}

	// Insert the message into the database
	timestamp := time.Now().Format(time.RFC3339)
	_, err = db.c.Exec(`
		INSERT INTO messages (id, conversationId, senderId, content, timestamp, forwardedMessageId, attachment)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, messageID, conversationID, senderID, content, timestamp, forwardedMessageID, attachment)
	if err != nil {
		return Message{}, fmt.Errorf("error saving message: %w", err)
	}

	// Return the saved message
	return Message{
		Id:               messageID,
		ConversationId:   conversationID,
		SenderId:         senderID,
		Content:          content,
		Timestamp:        timestamp,
		ForwardedMessage: forwardedMessageID,
		Attachment:       attachment,
	}, nil
}

func (db *appdbimpl) GetConversationMembers(conversationID string) ([]string, error) {
	rows, err := db.c.Query(`
		SELECT userId
		FROM conversation_members
		WHERE conversationId = ?
	`, conversationID)
	if err != nil {
		return nil, fmt.Errorf("error fetching conversation members: %w", err)
	}
	defer rows.Close()

	var members []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		members = append(members, userID)
	}
	return members, nil
}

func (db *appdbimpl) InsertDeliveryReceipt(messageID, userID, deliveredAt string) error {
	_, err := db.c.Exec(`
		INSERT INTO read_receipts (messageId, userId, deliveredAt)
		VALUES (?, ?, ?)
	`, messageID, userID, deliveredAt)
	if err != nil {
		return fmt.Errorf("error inserting delivery receipt: %w", err)
	}
	return nil
}

func (db *appdbimpl) IsUserInConversation(conversationID, userID string) (bool, error) {
	var exists bool
	err := db.c.QueryRow(`
		SELECT EXISTS(
			SELECT 1
			FROM conversation_members
			WHERE conversationId = ? AND userId = ?
		)
	`, conversationID, userID).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("error checking user membership: %w", err)
	}
	return exists, nil
}

func (db *appdbimpl) GetConversationDetails(conversationID string) (Conversation, error) {
	var conversation Conversation

	// Fetch basic conversation details
	err := db.c.QueryRow(`
		SELECT id, name, type, created_at
		FROM conversations
		WHERE id = ?
	`, conversationID).Scan(&conversation.Id, &conversation.Name, &conversation.Type, &conversation.CreatedAt)
	if err == sql.ErrNoRows {
		return Conversation{}, ErrConversationDoesNotExist
	}
	if err != nil {
		return Conversation{}, fmt.Errorf("error fetching conversation details: %w", err)
	}

	// Fetch members of the conversation
	members, err := db.GetConversationMembers(conversationID)
	if err != nil {
		return Conversation{}, fmt.Errorf("error fetching conversation members: %w", err)
	}
	conversation.Members = members

	// Fetch messages in the conversation
	messages, err := db.GetMessagesForConversation(conversationID)
	if err != nil {
		return Conversation{}, fmt.Errorf("error fetching conversation messages: %w", err)
	}
	conversation.Messages = messages

	return conversation, nil
}

func (db *appdbimpl) GetMessagesForConversation(conversationID string) ([]Message, error) {
	query := `
        SELECT 
            m.id, 
            m.conversationId, 
            m.senderId, 
            m.content, 
            m.timestamp, 
            m.forwardedMessageId, 
            m.attachment,
            u.name AS senderName,
            u.photo AS senderPhoto
        FROM 
            messages m
        JOIN 
            users u ON m.senderId = u.id
        WHERE 
            m.conversationId = ?
        ORDER BY 
            m.timestamp ASC
    `
	rows, err := db.c.Query(query, conversationID)
	if err != nil {
		return nil, fmt.Errorf("error fetching messages: %w", err)
	}
	defer rows.Close()
	var messages []Message
	for rows.Next() {
		var message Message
		var forwardedMessage sql.NullString
		var senderPhoto []byte
		err := rows.Scan(
			&message.Id,
			&message.ConversationId,
			&message.SenderId,
			&message.Content,
			&message.Timestamp,
			&forwardedMessage,
			&message.Attachment,
			&message.SenderName,
			&senderPhoto,
		)
		if err != nil {
			return nil, err
		}
		if senderPhoto != nil {
			message.SenderPhoto = base64.StdEncoding.EncodeToString(senderPhoto)
		} else {
			message.SenderPhoto = ""
		}
		if forwardedMessage.Valid {
			message.ForwardedMessage = &forwardedMessage.String
		}
		messages = append(messages, message)
	}
	return messages, nil
}

func (db *appdbimpl) GetMyConversations(userID string) ([]Conversation, error) {
	// SQL query to fetch conversations and the other member's photo for direct conversations
	query := `
	SELECT 
		c.id,
		CASE 
			WHEN c.type = 'direct' THEN 
				(SELECT u.name 
				 FROM users u 
				 JOIN conversation_members cm2 
				 ON u.id = cm2.userId 
				 WHERE cm2.conversationId = c.id AND u.id != ?)
			ELSE c.name
		END AS conversation_name,
		c.type,
		c.created_at,
		CASE 
			WHEN c.type = 'direct' THEN 
				(SELECT u.photo 
				 FROM users u 
				 JOIN conversation_members cm2 
				 ON u.id = cm2.userId 
				 WHERE cm2.conversationId = c.id AND u.id != ?)
			ELSE c.conversationPhoto
		END AS conversation_photo,
		(SELECT id FROM messages WHERE conversationId = c.id ORDER BY timestamp DESC LIMIT 1) AS last_message_id,
		(SELECT content FROM messages WHERE conversationId = c.id ORDER BY timestamp DESC LIMIT 1) AS last_message_content,
		(SELECT timestamp FROM messages WHERE conversationId = c.id ORDER BY timestamp DESC LIMIT 1) AS last_message_timestamp
	FROM conversations c
	JOIN conversation_members cm ON c.id = cm.conversationId
	WHERE cm.userId = ?
	ORDER BY last_message_timestamp DESC NULLS LAST;
	`
	// Execute the query
	rows, err := db.c.Query(query, userID, userID, userID)
	if err != nil {
		return nil, fmt.Errorf("error fetching conversations: %w", err)
	}
	defer rows.Close()

	// Slice to hold the conversations
	var conversations []Conversation

	// Iterate over the rows
	for rows.Next() {
		var conv Conversation
		var lastMessageID sql.NullString        // Handle nullable last_message_id
		var lastMessageContent sql.NullString   // Handle nullable last_message_content
		var lastMessageTimestamp sql.NullString // Handle nullable last_message_timestamp

		// Scan the row into the Conversation struct
		var convPhoto sql.NullString
		err := rows.Scan(
			&conv.Id,
			&conv.Name,
			&conv.Type,
			&conv.CreatedAt,
			&convPhoto,
			&lastMessageID,
			&lastMessageContent,
			&lastMessageTimestamp,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning conversation: %w", err)
		}

		if convPhoto.Valid {
			conv.ConversationPhoto.String = base64.StdEncoding.EncodeToString([]byte(convPhoto.String))
			conv.ConversationPhoto.Valid = true
		} else {
			conv.ConversationPhoto = sql.NullString{String: "", Valid: false}
		}

		if lastMessageID.Valid {
			conv.LastMessage = &Message{
				Id:        lastMessageID.String,
				Content:   lastMessageContent.String,
				Timestamp: lastMessageTimestamp.String,
			}
		}

		// Fetch conversation members
		members, err := db.GetConversationMembers(conv.Id)
		if err != nil {
			return nil, fmt.Errorf("error fetching conversation members: %w", err)
		}
		conv.Members = members

		// Append the conversation to the slice
		conversations = append(conversations, conv)
	}

	// Check for errors during iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return conversations, nil
}
