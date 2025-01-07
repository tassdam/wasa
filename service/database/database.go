package database

import (
	"database/sql"
	"errors"
	"fmt"
)

// Define errors similar to the other student's code, adjusted for WASAText domain
var ErrUserDoesNotExist = errors.New("User does not exist")
var ErrConversationDoesNotExist = errors.New("Conversation does not exist")
var ErrMessageDoesNotExist = errors.New("Message does not exist")
var ErrCommentDoesNotExist = errors.New("Comment does not exist")
var ErrGroupDoesNotExist = errors.New("Group does not exist")
var ErrGroupMemberDoesNotExist = errors.New("Group member does not exist")

// Define models as per your WASAText specification
type User struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type Conversation struct {
	Id          string    `json:"id"`
	Name        string    `json:"name"`
	Members     []string  `json:"members"`
	LastMessage *Message  `json:"lastMessage,omitempty"` // Add this line
	Messages    []Message `json:"messages,omitempty"`    // If you have full messages as well
}

type Message struct {
	Id               string  `json:"id"`
	ConversationId   string  `json:"conversationId"` // Add this line
	SenderId         string  `json:"senderId"`
	Content          string  `json:"content"`
	Timestamp        string  `json:"timestamp"`
	ForwardedMessage *string `json:"forwardedMessageId,omitempty"`
}

type Comment struct {
	Id        string `json:"id"`
	AuthorId  string `json:"authorId"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
}

type Group struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type GroupMember struct {
	UserId   string `json:"userId"`
	JoinedAt string `json:"joinedAt"`
}

// AppDatabase is the high-level interface for the DB operations.
// Actual CRUD methods will be defined in other files like user-db.go, message-db.go, etc.
type AppDatabase interface {
	Ping() error
	// ... other methods ...
	GetUserByName(name string) (User, error)
	CreateUser(u User) (User, error)
	UpdateUserName(userId string, newName string) (User, error)
	UpdateUserPhoto(userID string, photo []byte) error
}

// appdbimpl is the internal implementation of AppDatabase.
type appdbimpl struct {
	c *sql.DB
}

// New initializes the database and creates tables if they don’t exist, similar to the student’s code.
func New(db *sql.DB) (AppDatabase, error) {
	if db == nil {
		return nil, errors.New("database is required when building an AppDatabase")
	}

	// Enable foreign key support
	_, err := db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return nil, err
	}

	// Check if the primary table (users) exists to determine if schema creation is needed
	var tableName string
	err = db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='users';`).Scan(&tableName)
	if errors.Is(err, sql.ErrNoRows) {
		// Tables do not exist, create all necessary tables
		// Adjust these schemas based on your actual data requirements

		usersTable := `CREATE TABLE users (
			id TEXT NOT NULL PRIMARY KEY,
			name TEXT NOT NULL UNIQUE
		);`

		conversationsTable := `CREATE TABLE conversations (
			id TEXT NOT NULL PRIMARY KEY,
			name TEXT NOT NULL
		);`

		conversationMembersTable := `CREATE TABLE conversation_members (
			conversationId TEXT NOT NULL,
			userId TEXT NOT NULL,
			FOREIGN KEY (conversationId) REFERENCES conversations(id) ON DELETE CASCADE,
			FOREIGN KEY (userId) REFERENCES users(id) ON DELETE CASCADE,
			PRIMARY KEY(conversationId, userId)
		);`

		messagesTable := `CREATE TABLE messages (
			id TEXT NOT NULL PRIMARY KEY,
			conversationId TEXT NOT NULL,
			senderId TEXT NOT NULL,
			content TEXT NOT NULL,
			timestamp TEXT NOT NULL,
			forwardedMessageId TEXT,
			FOREIGN KEY (conversationId) REFERENCES conversations(id) ON DELETE CASCADE,
			FOREIGN KEY (senderId) REFERENCES users(id) ON DELETE CASCADE
		);`

		commentsTable := `CREATE TABLE comments (
			id TEXT NOT NULL PRIMARY KEY,
			messageId TEXT NOT NULL,
			authorId TEXT NOT NULL,
			content TEXT NOT NULL,
			timestamp TEXT NOT NULL,
			FOREIGN KEY (messageId) REFERENCES messages(id) ON DELETE CASCADE,
			FOREIGN KEY (authorId) REFERENCES users(id) ON DELETE CASCADE
		);`

		groupsTable := `CREATE TABLE groups (
			id TEXT NOT NULL PRIMARY KEY,
			name TEXT NOT NULL
		);`

		groupMembersTable := `CREATE TABLE group_members (
			groupId TEXT NOT NULL,
			userId TEXT NOT NULL,
			joinedAt TEXT NOT NULL,
			FOREIGN KEY (groupId) REFERENCES groups(id) ON DELETE CASCADE,
			FOREIGN KEY (userId) REFERENCES users(id) ON DELETE CASCADE,
			PRIMARY KEY (groupId, userId)
		);`

		// Execute the table creation statements
		creationQueries := []string{
			usersTable,
			conversationsTable,
			conversationMembersTable,
			messagesTable,
			commentsTable,
			groupsTable,
			groupMembersTable,
		}

		for _, q := range creationQueries {
			_, execErr := db.Exec(q)
			if execErr != nil {
				return nil, fmt.Errorf("error creating database structure: %w", execErr)
			}
		}
	} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
		// Some other error occurred when checking for tables
		return nil, err
	}

	return &appdbimpl{c: db}, nil
}

// Ping checks the connection to the database.
func (db *appdbimpl) Ping() error {
	return db.c.Ping()
}
