package database

import (
	"fmt"
)

func (db *appdbimpl) CommentMessage(commentID, messageID, authorID string) error {
	_, err := db.c.Exec("INSERT INTO comments (id, messageId, authorId) VALUES (?, ?, ?)", commentID, messageID, authorID)
	if err != nil {
		return fmt.Errorf("failed to insert comment: %w", err)
	}
	return nil
}

func (db *appdbimpl) UncommentMessage(messageID, authorID string) error {
	_, err := db.c.Exec("DELETE FROM comments WHERE messageId = ? AND authorId = ?", messageID, authorID)
	if err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}
	return nil
}
