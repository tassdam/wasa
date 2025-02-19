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
var ErrUnauthorizedToDeleteMessage = errors.New("Unauthorized To Delete Message")
var ErrGroupDoesNotExist = errors.New("Group does not exist")

type User struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Photo []byte `json:"photo,omitempty"`
}

type Group struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Photo []byte `json:"photo,omitempty"`
}

type Conversation struct {
	Id                string         `json:"id"`
	Name              string         `json:"name"`
	Type              string         `json:"type"`
	CreatedAt         string         `json:"createdAt"`
	Members           []string       `json:"members"`
	LastMessage       *Message       `json:"lastMessage,omitempty"`
	Messages          []Message      `json:"messages,omitempty"`
	ConversationPhoto sql.NullString `json:"conversationPhoto,omitempty"`
}

type Message struct {
	Id                string   `json:"id"`
	ConversationId    string   `json:"conversationId"`
	SenderId          string   `json:"senderId"`
	SenderName        string   `json:"senderName"`
	Content           string   `json:"content"`
	Timestamp         string   `json:"timestamp"`
	Attachment        []byte   `json:"attachment"`
	SenderPhoto       string   `json:"senderPhoto,omitempty"`
	ReactionCount     int      `json:"reactionCount"`
	ReactingUserNames []string `json:"reactingUserNames"`
	Status            string   `json:"status"`
	ReplyTo           string   `json:"replyTo,omitempty"`
	ReplyContent      string   `json:"replyContent,omitempty"`
	ReplySenderName   string   `json:"replySenderName,omitempty"`
	ReplyAttachment   []byte   `json:"replyAttachment,omitempty"`
}

type Comment struct {
	Id       string `json:"id"`
	AuthorId string `json:"authorId"`
}

type ReadReceipt struct {
	MessageId   string  `json:"messageId"`
	UserId      string  `json:"userId"`
	DeliveredAt string  `json:"deliveredAt"`
	ReadAt      *string `json:"readAt,omitempty"`
}

type AppDatabase interface {
	Ping() error
	GetUserByName(name string) (User, error)
	CreateUser(u User) (User, error)
	UpdateUserName(userId string, newName string) (User, error)
	UpdateUserPhoto(userID string, photo []byte) error
	SearchUsersByName(username string) ([]User, error)
	GetDirectConversation(senderID, recipientID string) (string, error)
	CreateDirectConversation(conversationID, senderID, recipientID string) error
	SaveMessage(conversationID, senderID, messageID, content string, attachment []byte, replyTo string) (Message, error)
	InsertDeliveryReceipt(messageID, userID, deliveredAt string) error
	IsUserInConversation(conversationID, userID string) (bool, error)
	GetConversationDetails(conversationID, currentUserID string) (Conversation, error)
	GetMessagesForConversation(conversationID string) ([]Message, error)
	GetMyConversations(userID string) ([]Conversation, error)
	GetConversationMembers(conversationID string) ([]string, error)
	GetUsersPhoto(userID string) (User, error)
	DeleteMessage(conversationID, messageID, userID string) error
	GetMessage(messageID, userID string) (Message, error)
	CreateGroupConversation(conversationID string, memberIDs []string, name string, photo []byte) error
	GetMyGroups(userID string) ([]Conversation, error)
	GetGroupInfo(groupID string) (Conversation, error)
	UpdateGroupName(groupId, newName string) error
	UpdateGroupPhoto(groupID string, photo []byte) error
	LeaveGroup(groupID, userID string) error
	AddUserToGroup(conversationID string, userID string) error
	CommentMessage(commentID, messageID, authorID string) error
	UncommentMessage(messageID, authorID string) error
	MarkMessagesAsRead(conversationID, userID string) error
}

type appdbimpl struct {
	c *sql.DB
}

func New(db *sql.DB) (AppDatabase, error) {
	if db == nil {
		return nil, errors.New("database is required when building an AppDatabase")
	}
	_, err := db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return nil, err
	}
	var tableName string
	err = db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='users';`).Scan(&tableName)
	if errors.Is(err, sql.ErrNoRows) {
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
			attachment BLOB,
			replyTo TEXT,  
			FOREIGN KEY (conversationId) REFERENCES conversations(id) ON DELETE CASCADE,
			FOREIGN KEY (senderId) REFERENCES users(id) ON DELETE CASCADE
		);`
		commentsTable := `CREATE TABLE comments (
			id TEXT NOT NULL PRIMARY KEY,
			messageId TEXT NOT NULL,
			authorId TEXT NOT NULL,
			UNIQUE(messageId, authorId),
			FOREIGN KEY (messageId) REFERENCES messages(id) ON DELETE CASCADE,
			FOREIGN KEY (authorId) REFERENCES users(id) ON DELETE CASCADE
		);`
		readReceiptsTable := `CREATE TABLE read_receipts (
			messageId TEXT NOT NULL,
			userId TEXT NOT NULL,
			deliveredAt TEXT NOT NULL,
			readAt TEXT, 
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
			readReceiptsTable,
		}
		for _, q := range creationQueries {
			_, execErr := db.Exec(q)
			if execErr != nil {
				return nil, fmt.Errorf("error creating database structure: %w", execErr)
			}
		}
	} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	return &appdbimpl{c: db}, nil
}

func (db *appdbimpl) Ping() error {
	return db.c.Ping()
}
