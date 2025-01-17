package database

import (
	"database/sql"
	"fmt"
)

// CreateUser inserts a new user into the database.
// Assumes u.Id is already set and unique, and u.Name is the username.
// Returns the user on success, or ErrUserDoesNotExist if something unexpected happens.
func (db *appdbimpl) CreateUser(u User) (User, error) {
	_, err := db.c.Exec("INSERT INTO users(id, name) VALUES (?, ?)", u.Id, u.Name)
	if err != nil {
		// Check if user already exists
		var existing User
		if errCheck := db.c.QueryRow("SELECT id, name FROM users WHERE name = ?", u.Name).Scan(&existing.Id, &existing.Name); errCheck != nil {
			if errCheck == sql.ErrNoRows {
				// Insertion failed and user not found - return the original error
				return u, err
			}
		}
		// User already exists, return that user
		return existing, nil
	}
	return u, nil
}

// GetUserByName fetches a user by their username (Name).
// Returns ErrUserDoesNotExist if no user matches.
func (db *appdbimpl) GetUserByName(name string) (User, error) {
	var u User
	if err := db.c.QueryRow("SELECT id, name FROM users WHERE name = ?", name).Scan(&u.Id, &u.Name); err != nil {
		if err == sql.ErrNoRows {
			return u, ErrUserDoesNotExist
		}
		return u, err
	}
	return u, nil
}

// GetUserById fetches a user by their unique id.
// Returns ErrUserDoesNotExist if no user matches.
func (db *appdbimpl) GetUserById(id string) (User, error) {
	var u User
	if err := db.c.QueryRow("SELECT id, name FROM users WHERE id = ?", id).Scan(&u.Id, &u.Name); err != nil {
		if err == sql.ErrNoRows {
			return u, ErrUserDoesNotExist
		}
		return u, err
	}
	return u, nil
}

// UpdateUserName updates the username of a user identified by userId.
// Returns the updated user or ErrUserDoesNotExist if no user is affected.
func (db *appdbimpl) UpdateUserName(userId, newName string) (User, error) {
	res, err := db.c.Exec(`UPDATE users SET name=? WHERE id=?`, newName, userId)
	if err != nil {
		return User{}, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return User{}, err
	} else if affected == 0 {
		return User{}, ErrUserDoesNotExist
	}

	// Return the updated user
	return db.GetUserById(userId)
}

func (db *appdbimpl) UpdateUserPhoto(userID string, photo []byte) error {
	// 1. Verify the user exists:
	var exists bool
	err := db.c.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE id=?)`, userID).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return ErrUserDoesNotExist
	}

	// 2. Update the user's photo column (must exist in your database schema)
	_, err = db.c.Exec(`UPDATE users SET photo=? WHERE id=?`, photo, userID)
	if err != nil {
		return err
	}

	return nil
}

// SearchUsersByName searches for users with usernames partially matching the input.
// SearchUsersByName searches for users by a partial match on the username
func (db *appdbimpl) SearchUsersByName(username string) ([]User, error) {
	var users []User
	rows, err := db.c.Query(`
        SELECT id, name, photo
        FROM users
        WHERE name LIKE ?`,
		"%"+username+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user User
		err := rows.Scan(&user.Id, &user.Name, &user.Photo)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	// Return an empty array if no results
	return users, nil
}

func (db *appdbimpl) GetUsersPhoto(userID string) (User, error) {
	var user User
	err := db.c.QueryRow(`
		SELECT id, name, photo 
		FROM users 
		WHERE id = ?
	`, userID).Scan(&user.Id, &user.Name, &user.Photo)
	if err == sql.ErrNoRows {
		return User{}, ErrUserDoesNotExist
	} else if err != nil {
		return User{}, fmt.Errorf("error fetching user by ID: %w", err)
	}
	return user, nil
}
