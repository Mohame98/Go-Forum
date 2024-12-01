package models

import (
	"golang.org/x/crypto/bcrypt"
	"database/sql"
	"errors"
	"time"
)

type User struct {
	ID             int
	User           string
	Email          string
	HashedPassword []byte
	Created        time.Time
}

type UserModel struct {
	DB *sql.DB
}

// NewUserModel creates a Users table and returns a new UserModel.
func NewUserModel(db *sql.DB) (*UserModel, error) {
	m := UserModel{db}
	err := m.createTable()
	if err != nil { return nil, err }
	return &m, nil
}

// createTable creates a Users table.
func (m *UserModel) createTable() error {
	stmt := 
	`
		CREATE TABLE IF NOT EXISTS users (
		uid INTEGER,
		user VARCHAR(255) UNIQUE NOT NULL,
		email VARCHAR(255) UNIQUE NOT NULL,
		hashed_password CHAR(60) NOT NULL,
		created DATETIME DEFAULT CURRENT_TIMESTAMP,

   	 	PRIMARY KEY (uid)
	);`
	_, err := m.DB.Exec(stmt)
	if err != nil {return err}
	return nil
}

// Insert inserts a user into the database.
func (m *UserModel) Insert(user, email, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil { return err }
	stmt := 
	`
		INSERT INTO users (user, email, hashed_password, created)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	`
	_, err = m.DB.Exec(stmt, user, email, string(hashedPassword))
	if err != nil { return err }
	return nil
}

func (m *UserModel) GetUserByID(id int) (*User, error) {
	stmt := 
	`
		SELECT uid, user, email, created
		FROM users
		WHERE uid = ?
	`
	row := m.DB.QueryRow(stmt, id)
	var u User

	err := row.Scan(&u.ID, &u.User, &u.Email, &u.Created)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {return nil, ErrNoRecord}
		return nil, err 
	}
	return &u, nil
}

func (m *UserModel) Authenticate(email, password string) (int, error) {
	var id int
	var hashedPassword []byte

	stmt := "SELECT uid, hashed_password FROM users WHERE email = ?"
	result := m.DB.QueryRow(stmt, email)
	err := result.Scan(&id, &hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrInvalidCredentials
		}
		return 0, err
	}
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, ErrInvalidCredentials
		}
		return 0, err
	}
	return id, nil
}

func (m *UserModel) CheckDuplicateEmail(registerEmail string) error {
	var email string
	stmt := "SELECT email FROM users WHERE email = ?"
	err := m.DB.QueryRow(stmt, registerEmail).Scan(&email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) { return nil }
		return err 
	}
	return ErrDuplicateEmail
}

func (m *UserModel) CheckDuplicateUser(registerUser string) error {
	var user string
	stmt := "SELECT user FROM users WHERE user = ?"
	err := m.DB.QueryRow(stmt, registerUser).Scan(&user)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) { return nil }
		return err 
	}
	return ErrDuplicateUser
}

func (m *UserModel) PasswordUpdate(uid int, currentPassword, newPassword string) error {
	var currentHashedPassword []byte
	stmt := "SELECT hashed_password FROM users WHERE uid = ?"

	err := m.DB.QueryRow(stmt, uid).Scan(&currentHashedPassword)
	if err != nil { return err }
	
	err = bcrypt.CompareHashAndPassword(currentHashedPassword, []byte(currentPassword))
	if err != nil { 
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrInvalidCredentials
		}
		return err
	}

	newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil { return err }
	
	stmt = "UPDATE users SET hashed_password = ? WHERE uid = ?"
	_, err = m.DB.Exec(stmt, string(newHashedPassword), uid)
	return err
}