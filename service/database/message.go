package database

import "time"

// Message represents a message in a conversation.
// ForwardedMsgID can be nil if the message is not a forwarded message.
type Message struct {
	ID             string
	ConversationID string
	SenderID       string
	Content        string
	Timestamp      time.Time
	ForwardedMsgID *string
}

// SendMessage inserts a new message into the messages table.
func (db *appdbimpl) SendMessage(m Message) error {
	_, err := db.c.Exec(
		`INSERT INTO messages (id, conversation_id, sender_id, content, timestamp, forwarded_message_id)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		m.ID, m.ConversationID, m.SenderID, m.Content, m.Timestamp, m.ForwardedMsgID,
	)
	return err
}

// DeleteMessage removes a message from the messages table.
func (db *appdbimpl) DeleteMessage(messageID string) error {
	_, err := db.c.Exec(`DELETE FROM messages WHERE id = ?`, messageID)
	return err
}

// ForwardMessage takes an existing message and creates a new message in another conversation,
// referencing the original via forwarded_message_id.
func (db *appdbimpl) ForwardMessage(originalMsgID, newMsgID, destinationConversationID, senderID string, timestamp time.Time) error {
	// First get the content of the original message
	var content string
	err := db.c.QueryRow(`SELECT content FROM messages WHERE id = ?`, originalMsgID).Scan(&content)
	if err != nil {
		return err
	}

	_, err = db.c.Exec(
		`INSERT INTO messages (id, conversation_id, sender_id, content, timestamp, forwarded_message_id)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		newMsgID, destinationConversationID, senderID, content, timestamp, originalMsgID,
	)
	return err
}

// GetMessagesByConversation returns all messages in a given conversation, possibly sorted by timestamp.
func (db *appdbimpl) GetMessagesByConversation(conversationID string) ([]Message, error) {
	rows, err := db.c.Query(`
        SELECT id, conversation_id, sender_id, content, timestamp, forwarded_message_id
        FROM messages
        WHERE conversation_id = ?
        ORDER BY timestamp DESC`, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []Message
	for rows.Next() {
		var m Message
		err = rows.Scan(&m.ID, &m.ConversationID, &m.SenderID, &m.Content, &m.Timestamp, &m.ForwardedMsgID)
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	return msgs, rows.Err()
}
