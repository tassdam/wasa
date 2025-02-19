package database

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"
)

func (db *appdbimpl) GetDirectConversation(senderID, recipientID string) (string, error) {
	var conversationID string
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
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("error checking conversation: %w", err)
	}
	return conversationID, nil
}

func (db *appdbimpl) CreateDirectConversation(conversationID, senderID, recipientID string) error {
	_, err := db.c.Exec(`
		INSERT INTO conversations (id, name, type, created_at, conversationPhoto)
		VALUES (?, '', 'direct', ?, '')
	`, conversationID, time.Now().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("error creating new conversation: %w", err)
	}
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
	conversationID, senderID, messageID, content string, attachment []byte, replyTo string,
) (Message, error) {
	var conversationExists bool
	err := db.c.QueryRow(`SELECT EXISTS(SELECT 1 FROM conversations WHERE id = ?)`, conversationID).Scan(&conversationExists)
	if err != nil {
		return Message{}, fmt.Errorf("error checking conversation existence: %w", err)
	}
	if !conversationExists {
		return Message{}, ErrConversationDoesNotExist
	}
	timestamp := time.Now().Format(time.RFC3339)
	_, err = db.c.Exec(`
        INSERT INTO messages (id, conversationId, senderId, content, timestamp, attachment, replyTo)
        VALUES (?, ?, ?, ?, ?, ?, ?)
    `, messageID, conversationID, senderID, content, timestamp, attachment, replyTo)
	if err != nil {
		return Message{}, fmt.Errorf("error saving message: %w", err)
	}
	return Message{
		Id:             messageID,
		ConversationId: conversationID,
		SenderId:       senderID,
		Content:        content,
		Timestamp:      timestamp,
		Attachment:     attachment,
		ReplyTo:        replyTo,
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
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating members: %w", err)
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

func (db *appdbimpl) GetConversationDetails(conversationID, currentUserID string) (Conversation, error) {
	var conversation Conversation
	var photoData []byte
	err := db.c.QueryRow(`
		SELECT id, name, type, created_at, conversationPhoto
		FROM conversations
		WHERE id = ?
	`, conversationID).Scan(
		&conversation.Id,
		&conversation.Name,
		&conversation.Type,
		&conversation.CreatedAt,
		&photoData,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Conversation{}, ErrConversationDoesNotExist
	}
	if err != nil {
		return Conversation{}, fmt.Errorf("error fetching conversation details: %w", err)
	}
	if len(photoData) > 0 {
		conversation.ConversationPhoto = sql.NullString{
			String: base64.StdEncoding.EncodeToString(photoData),
			Valid:  true,
		}
	} else {
		conversation.ConversationPhoto = sql.NullString{Valid: false}
	}
	members, err := db.GetConversationMembers(conversationID)
	if err != nil {
		return Conversation{}, fmt.Errorf("error fetching conversation members: %w", err)
	}
	conversation.Members = members
	if conversation.Type == "direct" {
		var otherUserID string
		for _, m := range members {
			if m != currentUserID {
				otherUserID = m
				break
			}
		}
		if otherUserID != "" {
			var userPhotoData []byte
			err := db.c.QueryRow("SELECT photo FROM users WHERE id = ?", otherUserID).Scan(&userPhotoData)
			if err == nil && len(userPhotoData) > 0 {
				conversation.ConversationPhoto = sql.NullString{
					String: base64.StdEncoding.EncodeToString(userPhotoData),
					Valid:  true,
				}
			}
		}
	}
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
    m.attachment,
    m.replyTo,
    u.name AS senderName,
    u.photo AS senderPhoto,
    ((SELECT COUNT(*) FROM conversation_members WHERE conversationId = m.conversationId) - 1) AS totalRecipients,
    (SELECT COUNT(*) FROM read_receipts WHERE messageId = m.id AND readAt IS NOT NULL) AS readCount,
    COUNT(c.id) AS reaction_count,
    GROUP_CONCAT(DISTINCT u2.name) AS reacting_user_names,
    IFNULL(r.content, '') AS replyContent,
    IFNULL(ru.name, '') AS replySenderName,
    r.attachment AS replyAttachment
FROM messages m
JOIN users u ON m.senderId = u.id
LEFT JOIN comments c ON m.id = c.messageId
LEFT JOIN users u2 ON c.authorId = u2.id
LEFT JOIN messages r ON m.replyTo = r.id
LEFT JOIN users ru ON r.senderId = ru.id
WHERE m.conversationId = ?
GROUP BY m.id
ORDER BY m.timestamp ASC;
`
	rows, err := db.c.Query(query, conversationID)
	if err != nil {
		return nil, fmt.Errorf("error fetching messages: %w", err)
	}
	defer rows.Close()
	var messages []Message
	for rows.Next() {
		var msg Message
		var senderPhoto []byte
		var totalRecipients, readCount, reactionCount int
		var reactingUserNames sql.NullString
		err := rows.Scan(
			&msg.Id,
			&msg.ConversationId,
			&msg.SenderId,
			&msg.Content,
			&msg.Timestamp,
			&msg.Attachment,
			&msg.ReplyTo,
			&msg.SenderName,
			&senderPhoto,
			&totalRecipients,
			&readCount,
			&reactionCount,
			&reactingUserNames,
			&msg.ReplyContent,
			&msg.ReplySenderName,
			&msg.ReplyAttachment,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning message row: %w", err)
		}
		if senderPhoto != nil {
			msg.SenderPhoto = base64.StdEncoding.EncodeToString(senderPhoto)
		}
		msg.ReactionCount = reactionCount
		if reactingUserNames.Valid && reactingUserNames.String != "" {
			msg.ReactingUserNames = strings.Split(reactingUserNames.String, ",")
		} else {
			msg.ReactingUserNames = []string{}
		}
		if totalRecipients > 0 && readCount >= totalRecipients {
			msg.Status = "✓✓"
		} else {
			msg.Status = "✓"
		}
		messages = append(messages, msg)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating message rows: %w", err)
	}
	return messages, nil
}

func (db *appdbimpl) GetMyConversations(userID string) ([]Conversation, error) {
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
		(SELECT m.id FROM messages m WHERE m.conversationId = c.id ORDER BY m.timestamp DESC LIMIT 1) AS last_message_id,
		(SELECT m.content FROM messages m WHERE m.conversationId = c.id ORDER BY m.timestamp DESC LIMIT 1) AS last_message_content,
		(SELECT m.timestamp FROM messages m WHERE m.conversationId = c.id ORDER BY m.timestamp DESC LIMIT 1) AS last_message_timestamp,
		(SELECT u.name FROM messages m 
		JOIN users u ON m.senderId = u.id 
		WHERE m.conversationId = c.id 
		ORDER BY m.timestamp DESC LIMIT 1) AS last_message_sender_name,
		(SELECT m.attachment FROM messages m   -- Fetch actual attachment data
		WHERE m.conversationId = c.id 
		ORDER BY m.timestamp DESC LIMIT 1) AS last_message_attachment
	FROM conversations c
	JOIN conversation_members cm ON c.id = cm.conversationId
	WHERE cm.userId = ?
	ORDER BY last_message_timestamp DESC NULLS LAST;
    `
	rows, err := db.c.Query(query, userID, userID, userID)
	if err != nil {
		return nil, fmt.Errorf("error fetching conversations: %w", err)
	}
	defer rows.Close()
	var conversations []Conversation
	for rows.Next() {
		var conv Conversation
		var (
			lastMessageID         sql.NullString
			lastMessageContent    sql.NullString
			lastMessageTimestamp  sql.NullString
			lastMessageSender     sql.NullString
			lastMessageAttachment []byte
			convPhoto             sql.NullString
		)
		err := rows.Scan(
			&conv.Id,
			&conv.Name,
			&conv.Type,
			&conv.CreatedAt,
			&convPhoto,
			&lastMessageID,
			&lastMessageContent,
			&lastMessageTimestamp,
			&lastMessageSender,
			&lastMessageAttachment,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning conversation: %w", err)
		}
		if convPhoto.Valid {
			conv.ConversationPhoto.String = base64.StdEncoding.EncodeToString([]byte(convPhoto.String))
			conv.ConversationPhoto.Valid = true
		}
		if lastMessageID.Valid {
			conv.LastMessage = &Message{
				Id:         lastMessageID.String,
				Content:    lastMessageContent.String,
				Timestamp:  lastMessageTimestamp.String,
				SenderName: lastMessageSender.String,
				Attachment: lastMessageAttachment,
			}
		}
		members, err := db.GetConversationMembers(conv.Id)
		if err != nil {
			return nil, fmt.Errorf("error fetching conversation members: %w", err)
		}
		conv.Members = members
		conversations = append(conversations, conv)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}
	return conversations, nil
}

func (db *appdbimpl) DeleteMessage(conversationID, messageID, userID string) error {
	var senderID string
	err := db.c.QueryRow(`
		SELECT senderId
		FROM messages
		WHERE conversationId = ? AND id = ?
	`, conversationID, messageID).Scan(&senderID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrMessageDoesNotExist
	}
	if err != nil {
		return fmt.Errorf("error fetching message: %w", err)
	}
	if senderID != userID {
		return ErrUnauthorizedToDeleteMessage
	}
	_, err = db.c.Exec(`
		DELETE FROM messages
		WHERE conversationId = ? AND id = ?
	`, conversationID, messageID)
	if err != nil {
		return fmt.Errorf("error deleting message: %w", err)
	}
	return nil
}

func (db *appdbimpl) GetMessage(messageID, userID string) (Message, error) {
	var message Message
	err := db.c.QueryRow(`
        SELECT 
            m.id, 
            m.conversationId, 
            m.senderId, 
            m.content, 
            m.timestamp, 
            m.attachment,
            u.name AS senderName
        FROM 
            messages m
        JOIN 
            users u ON m.senderId = u.id
        JOIN 
            conversation_members cm ON m.conversationId = cm.conversationId
        WHERE 
            m.id = ? AND cm.userId = ?
    `, messageID, userID).Scan(
		&message.Id,
		&message.ConversationId,
		&message.SenderId,
		&message.Content,
		&message.Timestamp,
		&message.Attachment,
		&message.SenderName,
	)
	if err == sql.ErrNoRows {
		return message, ErrMessageDoesNotExist
	}
	if err != nil {
		return message, fmt.Errorf("error fetching message: %w", err)
	}
	return message, nil
}

func (db *appdbimpl) MarkMessagesAsRead(conversationID, userID string) error {
	_, err := db.c.Exec(`
        UPDATE read_receipts
        SET readAt = CURRENT_TIMESTAMP
        WHERE messageId IN (SELECT id FROM messages WHERE conversationId = ?)
          AND userId = ?
          AND readAt IS NULL
    `, conversationID, userID)
	return err
}
