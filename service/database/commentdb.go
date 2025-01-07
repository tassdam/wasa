package database

// AddComment inserts a new comment into the comments table.
// It requires Comment.Id, Comment.AuthorId, Comment.Content, Comment.Timestamp, and the messageId to which the comment belongs.
// Returns the inserted comment or an error.
func (db *appdbimpl) AddComment(c Comment, messageId string) (Comment, error) {
	_, err := db.c.Exec(`
		INSERT INTO comments (id, messageId, authorId, content, timestamp)
		VALUES (?, ?, ?, ?, ?)`,
		c.Id, messageId, c.AuthorId, c.Content, c.Timestamp)
	if err != nil {
		return c, err
	}
	return c, nil
}

// RemoveComment deletes a comment by its id.
// If no rows are affected, it returns ErrCommentDoesNotExist.
// Depending on your rules, you might want to ensure that only the author of the comment can delete it.
// If so, you can pass `authorId` and add a WHERE clause for that as well.
func (db *appdbimpl) RemoveComment(commentId string, authorId string) error {
	res, err := db.c.Exec(`DELETE FROM comments WHERE id=? AND authorId=?`, commentId, authorId)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	} else if affected == 0 {
		return ErrCommentDoesNotExist
	}
	return nil
}

// (Optional) GetCommentsByMessageId retrieves all comments for a given message, if needed in the future.
// Not explicitly required by the endpoints you have, but useful if you need to display comments.
// Returns an empty slice if there are no comments.
func (db *appdbimpl) GetCommentsByMessageId(messageId string) ([]Comment, error) {
	var comments []Comment
	rows, err := db.c.Query(`
		SELECT id, authorId, content, timestamp
		FROM comments
		WHERE messageId=?
		ORDER BY timestamp ASC`, messageId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var c Comment
		err = rows.Scan(&c.Id, &c.AuthorId, &c.Content, &c.Timestamp)
		if err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return comments, nil
}
