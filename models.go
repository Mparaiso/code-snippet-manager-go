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
	Task    func(ctx context.Context) error `datastore:"-" json:"-"`
	Version int64
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

func (u User) GetID() int64                          { return u.ID }
func (u *User) SetID(id int64)                       { u.ID = id }
func (u User) GetVersion() int64                     { return u.Version }
func (u *User) SetVersion(version int64)             { u.Version = version }
func (u *User) SetCreated(date time.Time)            { u.Created = date }
func (u *User) SetUpdated(date time.Time)            { u.Updated = date }
func (u User) GetPassword() string                   { return u.Password }
func (u *User) SetPassword(password string)          { u.Password = password }
func (u *User) SetEncryptedPassword(password string) { u.EncryptedPassworld = password }

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

type Role struct {
	ID          int64
	Name        string
	Description string
	Version     int64
	Created     time.Time
	Updated     time.Time
	Locked      bool
}

func (r Role) GetID() int64               { return r.ID }
func (r *Role) SetID(id int64)            { r.ID = id }
func (r *Role) SetCreated(date time.Time) { r.Created = date }
func (r *Role) SetUpdated(date time.Time) { r.Updated = date }
func (r Role) GetVersion() int64          { return r.Version }
func (r *Role) SetVersion(version int64)  { r.Version = version }
func (r *Role) IsLocked() bool            { return r.Locked }

type Roles []*Role

func (roles Roles) GetByName(roleName string) *Role {
	for _, role := range roles {
		if role.Name == roleName {
			return role
		}

	}
	return nil
}

type UserRole struct {
	ID      int64
	RoleID  int64
	UserID  int64
	Created time.Time
	Updated time.Time
	Version int64
}

func (userRole UserRole) GetID() int64               { return userRole.ID }
func (userRole *UserRole) SetID(id int64)            { userRole.ID = id }
func (userRole *UserRole) SetCreated(date time.Time) { userRole.Created = date }
func (userRole *UserRole) SetUpdated(date time.Time) { userRole.Updated = date }
func (userRole UserRole) GetVersion() int64          { return userRole.Version }
func (userRole *UserRole) SetVersion(version int64)  { userRole.Version = version }

type Ancestor struct {
	ID string
}
