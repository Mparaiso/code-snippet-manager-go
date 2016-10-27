package smartsnippets

import (
	"fmt"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"
)

func MustParse(layout string, value string) time.Time {
	t, err := time.Parse(layout, value)
	if err != nil {
		panic(err)
	}
	return t
}

const Rfc2822 = "Mon, 02 Jan 2006 15:04:05 -0700"

func GetRootKey(ctx context.Context) *datastore.Key {
	return datastore.NewKey(ctx, "Ancestors", "Root", 0, nil)
}

// GetMigrations get al list of migrations
func GetMigrations() []*Migration {

	return []*Migration{
		{
			Created: MustParse(Rfc2822, "Wed, 26 Oct 2016 17:30:30 +0200"), Name: "000-root-ancestor", Task: func(ctx context.Context) error {

				rootKey := datastore.NewKey(ctx, "Ancestors", "Root", 0, nil)
				_, err := datastore.Put(ctx, rootKey, &Ancestor{ID: "Root"})
				return err
			}},
		{
			Name: "001-categories", Created: MustParse(Rfc2822, "Fri, 21 Oct 2016 09:11:26 +0200"), Task: func(ctx context.Context) error {
				categories := []*Category{
					{Title: "PHP", Description: "The PHP Language"},
					{Title: "Javascript", Description: "The Javascript Language"},
					{Title: "Go", Description: "The Go Language"},
					{Title: "Java", Description: "The Java Language"},
					{Title: "Ruby", Description: "The Ruby Language"},
					{Title: "Python", Description: "The Python Language"},
					{Title: "C", Description: "The C Language"},
					{Title: "C++", Description: "The C++ Language"},
					{Title: "SQL", Description: "The SQL Query Language"},
					{Title: "Scala", Description: "The Scala Language"},
					{Title: "Rust", Description: "The Rust Language"},
					{Title: "LISP", Description: "The LISP Language"},
					{Title: "HTML", Description: "The HTML Markup Language"},
					{Title: "XML", Description: "The XML Language"},
					{Title: "CSS", Description: "The CSS Language"},
					{Title: "Typescript", Description: "The Typescript Language"},
					{Title: "Swift", Description: "The Swift Language"},
					{Title: "Objective-C", Description: "The Objective-C Language"},
				}
				repository := NewCategoryRepository(ctx)
				for _, category := range categories {
					err := repository.Create(category)
					if err != nil {
						return fmt.Errorf("Error creating category %+v : %v", category, err)
					}
				}
				return nil
			},
		}, {
			Name: "002-Roles", Created: MustParse(Rfc2822, "Mon, 24 Oct 2016 04:12:36 +0200"), Task: func(ctx context.Context) error {
				roles := []*Role{
					{Name: "Root", Description: "The root administrators", Locked: true},
					{Name: "SuperAdmin", Description: "The Super Administrators", Locked: true},
					{Name: "User", Description: "Basic User", Locked: true},
					{Name: "Anonymous", Description: "Unauthenticated user, has no rights", Locked: true},
				}
				repository := NewRoleRepository(ctx)
				for _, role := range roles {
					err := repository.Create(role)
					if err != nil {
						return fmt.Errorf("Error creating role %+v : %v", role, err)
					}
				}

				return nil
			},
		}, {Name: "003-users", Created: MustParse("02 Jan 06 15:04 -0700", "18 Oct 16 08:49 +0200"), Task: func(ctx context.Context) error {
			user := &User{Nickname: "Anonymous"}
			return NewUserRepository(ctx).Create(user)

		}},
	}
}

// ExecuteMigrations execute all migrations
func ExecuteMigrations(ctx context.Context, migrations []*Migration) error {
	migrationRepository := NewMigrationRepository(ctx)
	for _, migration := range migrations {
		count, err := datastore.NewQuery(Kind.Migrations).Filter("Name =", migration.Name).Limit(1).Count(ctx)
		if err != nil {
			return err
		}
		if count != 0 {
			continue
		}
		err = migration.Task(ctx)
		if err != nil {
			return err
		}
		migrationRepository.Create(migration)
		if err != nil {
			return err
		}
	}
	return nil
}
