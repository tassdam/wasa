package database

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"time"
)

func (db *appdbimpl) CreateGroupConversation(conversationID string, memberIDs []string, name string, photo []byte) error {

	_, err := db.c.Exec(`
        INSERT INTO conversations (id, name, type, created_at, conversationPhoto)
        VALUES (?, ?, 'group', ?, ?)
    `, conversationID, name, time.Now().Format(time.RFC3339), photo)
	if err != nil {
		return fmt.Errorf("error creating new conversation: %w", err)
	}

	for _, memberID := range memberIDs {
		_, err = db.c.Exec(`
            INSERT INTO conversation_members (conversationId, userId)
            VALUES (?, ?)
        `, conversationID, memberID)
		if err != nil {
			return fmt.Errorf("error adding member %s to conversation_members: %w", memberID, err)
		}
	}

	return nil
}

func (db *appdbimpl) GetMyGroups(userID string) ([]Conversation, error) {
	query := `
    SELECT 
        c.id,
        c.name,
        c.conversationPhoto as photo
    FROM conversations c
    JOIN conversation_members cm ON c.id = cm.conversationId
    WHERE cm.userId = ? AND c.type = 'group'
    ORDER BY c.created_at DESC;
    `

	rows, err := db.c.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("error fetching groups: %w", err)
	}
	defer rows.Close()

	var groups []Conversation

	for rows.Next() {
		var group Conversation
		var photo sql.NullString

		err := rows.Scan(
			&group.Id,
			&group.Name,
			&photo,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning group: %w", err)
		}

		// Handle photo encoding same as conversations
		if photo.Valid {
			group.ConversationPhoto.String = base64.StdEncoding.EncodeToString([]byte(photo.String))
			group.ConversationPhoto.Valid = true
		} else {
			group.ConversationPhoto = sql.NullString{String: "", Valid: false}
		}

		groups = append(groups, group)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error after scanning groups: %w", err)
	}

	return groups, nil
}
