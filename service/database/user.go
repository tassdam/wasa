package database

import "errors"

// User represents a user entity in the system.
type User struct {
	ID   string
	Name string
}

// CreateUser inserts a new user into the `users` table.
func (db *appdbimpl) CreateUser(u User) error {
	_, err := db.c.Exec(
		"INSERT INTO users (id, name) VALUES (?, ?)",
		u.ID, u.Name,
	)
	return err
}

// GetUserByID retrieves a user from the `users` table by their ID.
func (db *appdbimpl) GetUserByID(userID string) (User, error) {
	var u User
	err := db.c.QueryRow(
		"SELECT id, name FROM users WHERE id = ?",
		userID,
	).Scan(&u.ID, &u.Name)
	return u, err
}

// UpdateUserName updates the name of a user identified by userID.
func (db *appdbimpl) UpdateUserName(userID, newName string) error {
	res, err := db.c.Exec(
		"UPDATE users SET name = ? WHERE id = ?",
		newName, userID,
	)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("no rows updated, user not found")
	}
	return nil
}
