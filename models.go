package smartsnippets // Migration is a DB migration
import (
	"time"

	"golang.org/x/net/context"
)

type Migration struct {
	ID      int64
	Name    string
	Created time.Time
	Updated time.Time
	Task    func(ctx context.Context) error `datastore:"-"`
}

func (m Migration) GetID() int64               { return m.ID }
func (m *Migration) SetID(id int64)            { m.ID = id }
func (m *Migration) SetCreated(date time.Time) { m.Created = date }
func (m *Migration) SetUpdated(date time.Time) { m.Updated = date }

// User is an app user
type User struct {
	ID                 int64
	Nickname           string
	Email              string
	Password           string
	EncryptedPassworld string
	Created            time.Time
	Updated            time.Time
	Version            int64
}

func (u User) GetID() int64               { return u.ID }
func (u *User) SetID(id int64)            { u.ID = id }
func (u User) GetVersion() int64          { return u.Version }
func (u *User) SetVersion(version int64)  { u.Version = version }
func (u *User) SetCreated(date time.Time) { u.Created = date }
func (u *User) SetUpdated(date time.Time) { u.Updated = date }

// Snippet is a code snippet
type Snippet struct {
	ID          int64
	Title       string
	Description string
	Content     string
	CategoryID  int64
	Category    *Category `datastore:"-"`
	Author      *User     `datastore:"-"`
	Created     time.Time
	Updated     time.Time
	Version     int64
}

func (s Snippet) GetID() int64               { return s.ID }
func (s *Snippet) SetID(id int64)            { s.ID = id }
func (s Snippet) GetVersion() int64          { return s.Version }
func (s *Snippet) SetVersion(version int64)  { s.Version = version }
func (s *Snippet) SetCreated(date time.Time) { s.Created = date }
func (s *Snippet) SetUpdated(date time.Time) { s.Updated = date }

// Category is a snippet category
type Category struct {
	ID          int64
	Title       string
	Description string
	Created     time.Time
	Updated     time.Time
	Version     int64
}

func (c Category) GetID() int64               { return c.ID }
func (c *Category) SetID(id int64)            { c.ID = id }
func (c *Category) SetCreated(date time.Time) { c.Created = date }
func (c *Category) SetUpdated(date time.Time) { c.Updated = date }
func (c Category) GetVersion() int64          { return c.Version }
func (c *Category) SetVersion(version int64)  { c.Version = version }

type Token struct {
	ID         int64
	UserID     int64
	Value      string
	Expiration time.Time
	Revoked    bool
	Created    time.Time
}
