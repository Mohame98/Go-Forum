package models

import (
	"database/sql"
	// "errors"
	"fmt"
	"time"
)

// Thread represents a single thread in the forum.
type Thread struct {
	ID      		int
	Title   		string
	Author 			*User
	Messages 		[]*Message
	ThreadCreated 	time.Time
}

// ThreadsModel is a model for interacting with the threads in the database.
type ThreadsModel struct {
	DB *sql.DB
}

// NewThreadModel creates a Threads table and returns a ThreadModel.
func NewThreadModel(db *sql.DB) (*ThreadsModel, error) {
	m := ThreadsModel{db}
	err := m.createTable()
	if err != nil {return nil,err}
	return &m, nil
}

// createTable creates a Threads table.
func (m *ThreadsModel) createTable() error {
	stmt :=
	`
		CREATE TABLE IF NOT EXISTS threads (
		tid INTEGER,
		title VARCHAR(100) NOT NULL,
		aid INTEGER NOT NULL,
		created DATETIME DEFAULT CURRENT_TIMESTAMP,

		PRIMARY KEY (tid),
		FOREIGN KEY (aid) REFERENCES users
	);`
	_, err := m.DB.Exec(stmt)
	if err != nil {return err}
	return nil
}

// Insert inserts a snippet into the database, and returns the snippet's id.
func (m *ThreadsModel) Insert(title string, authorID int) (int, error) {
	stmt := 
	`
		INSERT INTO threads (title, aid)
		VALUES (?, ?)
	`
	result, err := m.DB.Exec(stmt, title, authorID)
	if err != nil { return 0, fmt.Errorf("inserting thread into database: %w", err) }

	id, err := result.LastInsertId()
	if err != nil { return 0, fmt.Errorf("getting last insert ID: %w", err) }
	return int(id), nil
}

func (m *ThreadsModel) GetThreadByID(id int) (*Thread, error) {
	stmt := 
	`
		SELECT T.tid, T.title, T.created, U.uid, U.user, U.email
		FROM threads T, users U
		WHERE T.aid = U.uid AND T.tid = ?
	`
	row := m.DB.QueryRow(stmt, id)
	t, err := m.newThread(row, "ASC")
	if err != nil { return nil, err }
	return t, nil
}

// Latests retrieves the 10 latests threads from the database.
func (m *ThreadsModel) GetLatestThreads(limit, offset int) ([]*Thread, error) {
    stmt :=
    `
    SELECT T.tid, T.title, T.created, U.uid, U.user, U.email
    FROM threads T
    JOIN users U ON T.aid = U.uid
    ORDER BY T.created DESC
    LIMIT $1 OFFSET $2
    `
    rows, err := m.DB.Query(stmt, limit, offset)
	if err != nil { return nil, err }

	defer rows.Close()

	var threads []*Thread
	for rows.Next() {

		t, err := m.newThread(rows, "DESC")
		if err != nil { return nil, err }

		threads = append(threads, t)
	}
	if err = rows.Err(); err != nil { return nil, err }
	return threads, nil
}

// scanner implements the Scan function.
type scanner interface {
	Scan(dest ...any) error
}

// newThread creates a new Thread. It also creates a User to represent
// the Thread's author, and the Posts associated with that Thread.
func (m *ThreadsModel) newThread(s scanner, postOrder string) (*Thread, error) {
	var (
		t Thread
		u User
	)
	
	err := s.Scan(
		&t.ID, &t.Title, &t.ThreadCreated,
		&u.ID, &u.User, &u.Email,
	)

	if err != nil { return nil, err }
	t.Author = &u

	t.Messages, err = m.GetMsg(t.ID, postOrder)
	if err != nil { return nil, err }
	return &t, nil
}

// getPosts retrieves all Posts related to the Thread with the given threadID.
// The value of order must be "ASC" or "DESC".
func (m *ThreadsModel) GetMsg(threadID int, order string) ([]*Message, error) {
	stmt := fmt.Sprintf(
		`
			SELECT M.mid, M.body, M.created, U.uid, U.user, U.email
			FROM messages M, users U
			WHERE M.aid = U.uid AND M.tid = ?
			ORDER BY M.created %v
		`,
		order,
	)

	rows, err := m.DB.Query(stmt, threadID, order)
	if err != nil { return nil, err }

	defer rows.Close()
	var Messages []*Message
	for rows.Next() {

		var (
			m Message
			u User
		)

		err := rows.Scan(
			&m.ID, &m.Body, &m.MessageCreated,
			&u.ID, &u.User, &u.Email,
		)

		if err != nil { return nil, err }
		m.Author = &u
		Messages = append(Messages, &m)
	}
	if err = rows.Err(); err != nil { return nil, err}
	return Messages, nil
}

func (m *ThreadsModel) CountThreads() (int, error) {
    var count int
    err := m.DB.QueryRow(`SELECT COUNT(*) FROM threads`).Scan(&count)
    if err != nil {
        return 0, err
    }
    return count, nil
}