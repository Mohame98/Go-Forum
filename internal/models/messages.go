package models

import (
	"database/sql"
	"time"
	"fmt"
)

// Message represents a single message in a forum thread.
type Message struct {
	ID      		int
	Author			*User
	Body			string
	MessageCreated 	time.Time
}

// MessagesModel is a model for interacting with the messages in the database.
type MessagesModel struct {
	DB *sql.DB
}

// NewPostModel creates a Posts table and returns a new PostModel.
func NewMessagesModel(db *sql.DB) (*MessagesModel, error) {
	m := MessagesModel{db}
	err := m.createTable()
	if err != nil { return nil, err }
	return &m, nil
}

// createTable creates a Posts table.
func (m *MessagesModel) createTable() error {
	stmt :=
	`
		CREATE TABLE IF NOT EXISTS messages (
		mid INTEGER,
		tid INTEGER NOT NULL,
		aid INTEGER NOT NULL,
		body VARCHAR(1000) NOT NULL,
		created DATETIME DEFAULT CURRENT_TIMESTAMP,

		PRIMARY KEY (mid),
		FOREIGN KEY (tid) REFERENCES threads,
		FOREIGN KEY (aid) REFERENCES users
	);`
	_, err := m.DB.Exec(stmt)
	if err != nil { return err}
	return nil
}

// Insert inserts a snippet into the database, and returns the snippet's id.
func (m *MessagesModel) Insert(threadID int, authorID int, body string) (int, error) {
	stmt := 
	`
		INSERT INTO messages (tid, aid, body)
		VALUES (?, ?, ?)
	`
	result, err := m.DB.Exec(stmt, threadID, authorID, body)
	if err != nil { return 0, fmt.Errorf("inserting thread into database: %w", err) }

	id, err := result.LastInsertId()
	if err != nil { return 0, fmt.Errorf("getting last insert ID: %w", err) }
	return int(id), nil
}