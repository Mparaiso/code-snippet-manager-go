package smartsnippets_test

import (
	"testing"

	"github.com/Mparaiso/expect-go"
	app "github.com/Mparaiso/snipped-go"
	"golang.org/x/net/context"
	"google.golang.org/appengine/aetest"
)

func TestRole_FindBy(t *testing.T) {
	context, done, err := aetest.NewContext()
	expect.Expect(t, err, nil)
	defer done()

	err = app.ExecuteMigrations(context, app.GetMigrations())
	expect.Expect(t, err, nil)
	repository := app.NewRoleRepository(context)
	var roles []*app.Role
	err = repository.FindBy(app.Query{Query: map[string]interface{}{"Name=": "User"}}, &roles)
	expect.Expect(t, err, nil)
	expect.Expect(t, len(roles), 1)
	SubTestUserRepositoryCreate(t, context)
}

func SubTestUserRepositoryCreate(t *testing.T, ctx context.Context) {
	repository := app.NewUserRepository(ctx)
	user := &app.User{Nickname: "JohnDoe", Email: "john.doe@acme.com", Password: "Password"}
	err := repository.Create(user)
	expect.Expect(t, err, nil)
	expect.Expect(t, user.GetID() > 0, true)
}
