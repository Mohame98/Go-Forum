package models

import (
	"database/sql"
)

// NewModels creates all models necessary for the application.
func NewModels(db *sql.DB) (*ThreadsModel, *UserModel, *MessagesModel, error) {
	threadModel, err := NewThreadModel(db)
	if err != nil { return nil, nil, nil, err }
	
	userModel, err := NewUserModel(db)
	if err != nil { return nil, nil, nil, err }

	postModel, err := NewMessagesModel(db)
	if err != nil { return nil, nil, nil, err }

	return threadModel, userModel, postModel, nil
}