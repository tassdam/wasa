package database

// Conversation represents a conversation in the system.
// Adjust fields according to your actual schema.
type Conversation struct {
	ID      string
	Name    string
	Members []string // Retrieved from a join table
	// You can add more fields as needed, like LastMessageID, etc.
}

// ListConversations returns the conversations a user is part of.
func (db *appdbimpl) ListConversations(userID string) ([]Conversation, error) {
	rows, err := db.c.Query(`
        SELECT c.id, c.name
        FROM conversations c
        INNER JOIN conversation_members cm ON c.id = cm.conversation_id
        WHERE cm.user_id = ?`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var convs []Conversation
	for rows.Next() {
		var conv Conversation
		if err := rows.Scan(&conv.ID, &conv.Name); err != nil {
			return nil, err
		}
		// Fetch members for each conversation
		members, err := db.getConversationMembers(conv.ID)
		if err != nil {
			return nil, err
		}
		conv.Members = members
		convs = append(convs, conv)
	}
	return convs, rows.Err()
}

// getConversationMembers is a helper function to retrieve members of a conversation
func (db *appdbimpl) getConversationMembers(conversationID string) ([]string, error) {
	rows, err := db.c.Query(`
        SELECT user_id
        FROM conversation_members
        WHERE conversation_id = ?`, conversationID)
	if err != nil {
		return nil, err
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
	return members, rows.Err()
}

// GetConversationDetails retrieves all details of a specific conversation,
// including its members and possibly its messages.
func (db *appdbimpl) GetConversationDetails(conversationID string) (Conversation, error) {
	var conv Conversation
	err := db.c.QueryRow(`
        SELECT id, name
        FROM conversations
        WHERE id = ?`, conversationID).Scan(&conv.ID, &conv.Name)
	if err != nil {
		return conv, err
	}

	members, err := db.getConversationMembers(conversationID)
	if err != nil {
		return conv, err
	}
	conv.Members = members

	return conv, nil
}

// AddUserToConversation adds a user to a conversation via the conversation_members table.
func (db *appdbimpl) AddUserToConversation(conversationID, userID string) error {
	_, err := db.c.Exec(`
        INSERT INTO conversation_members (conversation_id, user_id) VALUES (?, ?)`,
		conversationID, userID)
	return err
}

// CreateConversation creates a new conversation and optionally adds initial members.
func (db *appdbimpl) CreateConversation(conv Conversation) error {
	_, err := db.c.Exec(`INSERT INTO conversations (id, name) VALUES (?, ?)`,
		conv.ID, conv.Name)
	if err != nil {
		return err
	}

	// Add members if provided
	for _, m := range conv.Members {
		if err := db.AddUserToConversation(conv.ID, m); err != nil {
			return err
		}
	}
	return nil
}

// UpdateConversationName updates the name of a conversation.
func (db *appdbimpl) UpdateConversationName(conversationID, newName string) error {
	_, err := db.c.Exec(`UPDATE conversations SET name = ? WHERE id = ?`,
		newName, conversationID)
	return err
}
