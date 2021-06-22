package database

import (
	"database/sql"
	"fmt"

	"github.com/gregoryalbouy/goshrink/internal"
	"github.com/jmoiron/sqlx"
)

// Ensure service implements interface.
var _ internal.UserService = (*userService)(nil)

// UserService represents a service for managing users, it implements the interface internal.UserService.
type userService struct {
	db *sqlx.DB
}

// NewUserService returns a `userService` that implements
// `internal.UserService` interface.
func NewUserService(db *DB) internal.UserService {
	return &userService{db: db.sqlx}
}

// FindByID retrieves a user by its ID.
func (s *userService) FindByID(userID int) (internal.User, error) {
	return s.find("id", userID)
}

// FindOne retrieves a user by its username.
func (s *userService) FindByUsername(username string) (internal.User, error) {
	return s.find("username", username)
}

func (s *userService) find(col string, val interface{}) (internal.User, error) {
	u := internal.User{}

	var acceptableCol = map[string]bool{
		"username": true,
		"id":       true,
	}
	if _, ok := acceptableCol[col]; !ok {
		return u, fmt.Errorf("illegal column %s", col)
	}

	if err := s.db.Get(
		&u,
		fmt.Sprintf("SELECT * FROM V_user_avatar WHERE %s = ?", col),
		val,
	); err != nil {
		return internal.User{}, err
	}
	return u, nil
}

// FindCreds retrieves a user credentials by its username.
func (s *userService) FindCreds(username string) (internal.User, error) {
	u := internal.User{}

	if err := s.db.Get(
		&u,
		"SELECT id, username, password FROM user WHERE username = ?",
		username,
	); err != nil {
		return internal.User{}, err
	}

	return u, nil
}

func (s *userService) SetAvatarURL(userID int, url string) error {
	if _, err := s.db.Exec(
		"REPLACE INTO avatar (user_id, avatar_url) VALUES (?, ?)",
		userID, url,
	); err != nil {
		return err
	}

	return nil
}

// InsertOne inserts a user in the database.
func (s *userService) InsertOne(u internal.User) error {
	// Compute email as NULL if it is an empty string
	email := sql.NullString{}
	if u.Email == "" {
		email.String = u.Email
		email.Valid = true
	}

	// Insert user
	if _, err := s.db.Exec(
		"INSERT INTO user (username, email, password) VALUES (?, ?, ?)",
		u.Username, email, u.Password,
	); err != nil {
		return err
	}

	// Set avatar URL if they have one
	if u.AvatarURL != "" {
		return s.SetAvatarURL(u.ID, u.AvatarURL)
	}

	return nil
}
