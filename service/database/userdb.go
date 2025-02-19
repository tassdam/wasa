package database

import (
	"database/sql"
	"fmt"
)

func (db *appdbimpl) CreateUser(u User) (User, error) {
	_, err := db.c.Exec("INSERT INTO users(id, name, photo) VALUES (?, ?, ?)", u.Id, u.Name, u.Photo)
	if err != nil {
		var existing User
		if errCheck := db.c.QueryRow("SELECT id, name FROM users WHERE name = ?", u.Name).Scan(&existing.Id, &existing.Name); errCheck != nil {
			if errCheck == sql.ErrNoRows {
				return u, err
			}
		}
		return existing, nil
	}
	return u, nil
}

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
	return db.GetUserById(userId)
}

func (db *appdbimpl) UpdateUserPhoto(userID string, photo []byte) error {
	var exists bool
	err := db.c.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE id=?)`, userID).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return ErrUserDoesNotExist
	}
	_, err = db.c.Exec(`UPDATE users SET photo=? WHERE id=?`, photo, userID)
	if err != nil {
		return err
	}
	return nil
}

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
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user results: %w", err)
	}
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
