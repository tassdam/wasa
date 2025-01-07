package database

import (
	"database/sql"
)

// SendMessage inserts a new message into the messages table.
// Expects m.Id, m.ConversationId, m.SenderId, m.Content, m.Timestamp to be set.
// forwardedMessageId is optional (nil if not forwarding).
// Returns the inserted message or an error.
func (db *appdbimpl) SendMessage(m Message) (Message, error) {
	_, err := db.c.Exec(`
		INSERT INTO messages (id, conversationId, senderId, content, timestamp, forwardedMessageId)
		VALUES (?, ?, ?, ?, ?, ?)`,
		m.Id, m.ConversationId, m.SenderId, m.Content, m.Timestamp, m.ForwardedMessage)
	if err != nil {
		return m, err
	}
	return m, nil
}

// ForwardMessage takes an existing message and creates a new one with the same content,
// referencing the original via forwardedMessageId. The caller should have looked up the content
// of the original message first or pass it in after retrieval. If you want this function to
// handle retrieval, adjust accordingly.
//
// For simplicity, assume the caller provides newMsg with the Id, destinationConversationId, senderId, and timestamp.
// The ForwardedMessage field is set to the original message ID, and the content is copied from the original.
//
// If the original message does not exist, return an error.
func (db *appdbimpl) ForwardMessage(originalMsgId string, newMsg Message) (Message, error) {
	// First, fetch the content of the original message to be forwarded
	var origContent string
	if err := db.c.QueryRow(`
		SELECT content FROM messages WHERE id = ?`,
		originalMsgId).Scan(&origContent); err != nil {
		if err == sql.ErrNoRows {
			return newMsg, ErrMessageDoesNotExist
		}
		return newMsg, err
	}

	// Now insert the new forwarded message
	newMsg.ForwardedMessage = &originalMsgId
	newMsg.Content = origContent

	_, err := db.c.Exec(`
		INSERT INTO messages (id, conversationId, senderId, content, timestamp, forwardedMessageId)
		VALUES (?, ?, ?, ?, ?, ?)`,
		newMsg.Id, newMsg.ConversationId, newMsg.SenderId, newMsg.Content, newMsg.Timestamp, newMsg.ForwardedMessage)
	if err != nil {
		return newMsg, err
	}
	return newMsg, nil
}

// DeleteMessage removes a message by its id. This should also cascade-delete comments if foreign keys and ON DELETE CASCADE are set up.
// Returns ErrMessageDoesNotExist if no rows are deleted.
func (db *appdbimpl) DeleteMessage(messageId string) error {
	res, err := db.c.Exec(`DELETE FROM messages WHERE id=?`, messageId)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	} else if affected == 0 {
		return ErrMessageDoesNotExist
	}
	return nil
}

// (Optional) GetMessageById could be useful if you need to verify existence or read details:
func (db *appdbimpl) GetMessageById(messageId string) (Message, error) {
	var msg Message
	err := db.c.QueryRow(`
		SELECT id, senderId, content, timestamp, forwardedMessageId, conversationId
		FROM messages
		WHERE id=?`,
		messageId).Scan(&msg.Id, &msg.SenderId, &msg.Content, &msg.Timestamp, &msg.ForwardedMessage, &msg.ConversationId)
	if err != nil {
		if err == sql.ErrNoRows {
			return msg, ErrMessageDoesNotExist
		}
		return msg, err
	}
	return msg, nil
}
