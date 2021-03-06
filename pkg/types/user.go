package types

import (
	"fmt"
	"time"

	"github.com/blend/go-sdk/oauth"
	"github.com/blend/go-sdk/uuid"
)

// ApplyProfileToUser applies an oauth proflie.
func ApplyProfileToUser(u *User, p oauth.Profile) {
	u.Email = p.Email
	u.GivenName = p.GivenName
	u.FamilyName = p.FamilyName
	u.Locale = p.Locale
	u.PictureURL = p.PictureURL
}

// User is a user
type User struct {
	ID           uuid.UUID `db:"id,pk"`
	CreatedUTC   time.Time `db:"created_utc"`
	LastLoginUTC time.Time `db:"last_login_utc"`
	LastSeenUTC  time.Time `db:"last_seen_utc"`
	ProfileID    string    `db:"profile_id"`
	GivenName    string    `db:"given_name"`
	FamilyName   string    `db:"family_name"`
	PictureURL   string    `db:"picture_url"`
	Locale       string    `db:"locale"`
	Email        string    `db:"email"`
}

// TableName returns the mapped table name.
func (u User) TableName() string { return "users" }

// NewTestUser returns a test user.
func NewTestUser() User {
	return User{
		ID:           uuid.V4(),
		CreatedUTC:   time.Now().UTC().Add(-time.Hour),
		LastLoginUTC: time.Now().UTC().Add(-time.Minute),
		LastSeenUTC:  time.Now().UTC().Add(-time.Second),
		ProfileID:    uuid.V4().String(),
		GivenName:    uuid.V4().String(),
		FamilyName:   uuid.V4().String(),
		PictureURL:   fmt.Sprintf("https://example.com/%s.jpg", uuid.V4().String()),
		Locale:       "en-us",
		Email:        fmt.Sprintf("%s@example.com", uuid.V4().String()),
	}
}
