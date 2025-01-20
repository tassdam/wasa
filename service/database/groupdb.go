package database

import (
	"fmt"
	"time"
)

func (db *appdbimpl) CreateGroupConversation(conversationID string, senderID string, recipientIDs []string, name string, photo []byte) error {

	_, err := db.c.Exec(`
        INSERT INTO conversations (id, name, type, created_at, conversationPhoto)
        VALUES (?, ?, 'group', ?, ?)
    `, conversationID, name, time.Now().Format(time.RFC3339), photo)
	if err != nil {
		return fmt.Errorf("error creating new conversation: %w", err)
	}

	_, err = db.c.Exec(`
        INSERT INTO conversation_members (conversationId, userId, isAdmin)
        VALUES (?, ?, True)
    `, conversationID, senderID)
	if err != nil {
		return fmt.Errorf("error adding sender to conversation_members: %w", err)
	}

	for _, recipientID := range recipientIDs {
		_, err = db.c.Exec(`
            INSERT INTO conversation_members (conversationId, userId, isAdmin)
            VALUES (?, ?, False)
        `, conversationID, recipientID)
		if err != nil {
			return fmt.Errorf("error adding recipient %s to conversation_members: %w", recipientID, err)
		}
	}

	return nil
}
