package database

import (
	"database/sql"
	"errors"
	"fmt"
)

var ErrUserDoesNotExist = errors.New("User does not exist")
var ErrConversationDoesNotExist = errors.New("Conversation does not exist")
var ErrMessageDoesNotExist = errors.New("Message does not exist")
var ErrCommentDoesNotExist = errors.New("Comment does not exist")

type User struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Photo []byte `json:"photo,omitempty"`
}

type Conversation struct {
	Id                string         `json:"id"`
	Name              string         `json:"name"`
	Type              string         `json:"type"`      // 'direct' or 'group'
	CreatedAt         string         `json:"createdAt"` // Timestamp of conversation creation
	Members           []string       `json:"members"`
	LastMessage       *Message       `json:"lastMessage,omitempty"`
	Messages          []Message      `json:"messages,omitempty"`
	ConversationPhoto sql.NullString `json:"conversationPhoto,omitempty"`
}

type Message struct {
	Id               string  `json:"id"`
	ConversationId   string  `json:"conversationId"`
	SenderId         string  `json:"senderId"`
	Content          string  `json:"content"`
	Timestamp        string  `json:"timestamp"`
	ForwardedMessage *string `json:"forwardedMessageId,omitempty"`
	Attachment       []byte  `json:"attachment,omitempty"` // Binary data for photos or GIFs
}

type Comment struct {
	Id        string `json:"id"`
	AuthorId  string `json:"authorId"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
}

type ReadReceipt struct {
	MessageId   string  `json:"messageId"`
	UserId      string  `json:"userId"`
	DeliveredAt string  `json:"deliveredAt"`
	ReadAt      *string `json:"readAt,omitempty"` // Nullable for messages not yet read
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
	SearchUsersByName(username string) ([]User, error)
	GetDirectConversation(senderID, recipientID string) (string, error)
	CreateDirectConversation(conversationID, senderID, recipientID string) error
	SaveMessage(conversationID, senderID, messageID, content string, forwardedMessageID *string, attachment []byte) (Message, error)
	InsertDeliveryReceipt(messageID, userID, deliveredAt string) error
	IsUserInConversation(conversationID, userID string) (bool, error)
	GetConversationDetails(conversationID string) (Conversation, error)
	GetMessagesForConversation(conversationID string) ([]Message, error)
	GetMyConversations(userID string) ([]Conversation, error)
	GetConversationMembers(conversationID string) ([]string, error)
	GetUsersPhoto(userID string) (User, error)
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
			name TEXT NOT NULL UNIQUE,
			photo BLOB
		);`

		conversationsTable := `CREATE TABLE conversations (
			id TEXT NOT NULL PRIMARY KEY,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			created_at TEXT NOT NULL,
			conversationPhoto BLOB
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
			attachment BLOB,
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

		readReceiptsTable := `CREATE TABLE read_receipts (
			messageId TEXT NOT NULL,
			userId TEXT NOT NULL,
			deliveredAt TEXT NOT NULL,
			readAt TEXT, -- Nullable
			PRIMARY KEY (messageId, userId),
			FOREIGN KEY (messageId) REFERENCES messages(id) ON DELETE CASCADE,
			FOREIGN KEY (userId) REFERENCES users(id) ON DELETE CASCADE
		);`

		creationQueries := []string{
			usersTable,
			conversationsTable,
			conversationMembersTable,
			messagesTable,
			commentsTable,
			readReceiptsTable, // Add the read_receipts table
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
