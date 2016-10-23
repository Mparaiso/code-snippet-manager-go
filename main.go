package smartsnippets

import (
	"fmt"
	"net/http"
	"reflect"
	"time"

	"golang.org/x/net/context"

	"sync"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	matcher "github.com/Mparaiso/request-matcher-go"
	m "github.com/Mparaiso/simple-middleware-go"
)

// Kind list app kinds
var Kind = struct{ Users, Migrations, Snippets, Categories string }{"Users", "Migrations", "Snippets", "Categories"}

func init() {
	http.Handle("/", NewApp().Compile())
}

var (
	_ matcher.MatcherProvider = new(Route)
	_ EndPointContainer       = new(DefaultEndPointContainer)
)

type UserEndPointContainerFactory struct {
}

func (UserEndPointContainerFactory) Create(container m.Container) EndPointContainer {
	return &DefaultEndPointContainer{
		ContextAwareContainer: container.(*Container),
		kind:      Kind.Users,
		prototype: reflect.TypeOf(User{}),
	}
}

type SnippetEndPointContainerFactory struct{}

func (SnippetEndPointContainerFactory) Create(container m.Container) EndPointContainer {
	return &DefaultEndPointContainer{
		ContextAwareContainer: container.(*Container),
		kind:      Kind.Snippets,
		prototype: reflect.TypeOf(Snippet{}),
	}
}

type CategoryEndPointContainerFactory struct{}

func (CategoryEndPointContainerFactory) Create(container m.Container) EndPointContainer {
	return &DefaultEndPointContainer{
		ContextAwareContainer: container.(*Container),
		kind:      Kind.Categories,
		prototype: reflect.TypeOf(Category{}),
	}
}

type ContextAwareContainer interface {
	m.Container
	ContextProvider
}

type DefaultEndPointContainer struct {
	kind      string
	prototype reflect.Type
	ContextAwareContainer
	repository Repository
}

func NewDefaultEndPointContainer(kind string,
	prototype reflect.Type,
	container ContextAwareContainer,
) *DefaultEndPointContainer {
	return &DefaultEndPointContainer{kind: kind, prototype: prototype, ContextAwareContainer: container}
}

func (endpointContainer *DefaultEndPointContainer) GetRepository() Repository {
	if endpointContainer.repository == nil {
		endpointContainer.repository = NewAppengineRepositoryProvider(endpointContainer, endpointContainer.kind)
	}
	return endpointContainer.repository
}

func (endpointContainer DefaultEndPointContainer) GetPrototype() reflect.Type {
	return endpointContainer.prototype
}

// App is the web application
type App struct {
	*sync.Once
	Debug bool
	*Router
}

func (a App) GetContainer(w http.ResponseWriter, r *http.Request) m.Container {
	return NewContainer(w, r)
}

func NewApp() *App {
	app := new(App)
	app.Debug = true
	app.Router = NewRouter()
	app.Once = new(sync.Once)

	app.ContainerFactory = app

	snippetEndpoint := NewEndpoint(new(SnippetEndPointContainerFactory))
	categoryEndpoint := NewEndpoint(new(CategoryEndPointContainerFactory))
	userEndpoint := NewEndpoint(new(UserEndPointContainerFactory))

	app.Use(func(next m.Handler) m.Handler {
		return func(c m.Container) {
			container := c.(*Container)
			container.SetContainerOptions(ContainerOptions{Debug: app.Debug})
			app.Do(func() {
				ctx := container.GetContext()
				if err := ExecuteMigrations(ctx, GetMigrations()); err != nil {
					log.Errorf(ctx, "Error during migration '%s'.", err.Error())
				} else {
					log.Infof(ctx, "Migrations done")
				}
			})
			next(c)
		}
	}).
		Get("/", index).
		Mount("/snippets", snippetEndpoint).
		Mount("/categories", categoryEndpoint).
		Mount("/users", userEndpoint)
	return app
}

type AppengineRepositoryProvider struct {
	Kind string
	ContextProvider
	Repository
	listeners []Listener
}

func NewAppengineRepositoryProvider(contextProvider ContextProvider, kind string, listeners ...Listener) *AppengineRepositoryProvider {
	return &AppengineRepositoryProvider{Kind: kind, ContextProvider: contextProvider, listeners: listeners}
}

func (provider *AppengineRepositoryProvider) GetRepository() (Repository, error) {
	if provider.Repository == nil {
		provider.Repository = NewDefaultRepository(provider.GetContext(), provider.Kind, provider.listeners...)
	}
	return provider.Repository, nil
}

func index(c m.Container) {
	fmt.Fprint(c.GetResponseWriter(), "Hello Smart Snippets")
}

type ContainerOptions struct {
	Debug bool
}

// Container is a container
type Container struct {
	m.Container
	ContextFactory ContextFactory
	context.Context
	containerOptions ContainerOptions
	logger           m.Logger
}

func (c Container) IsDebug() bool {
	return c.containerOptions.Debug
}

// NewContainer creates a new container
func NewContainer(w http.ResponseWriter, r *http.Request) *Container {
	return &Container{Container: m.DefaultContainer{Request: r, ResponseWriter: w}}
}
func (c *Container) SetContainerOptions(options ContainerOptions) {
	c.containerOptions = options
}
func (c *Container) Error(err error, statusCode int) {
	if c.IsDebug() {
		c.Container.Error(err, statusCode)
	} else {
		c.Container.Error(m.StatusError(statusCode), statusCode)
		c.MustGetLogger().Log(m.Error, err)
	}
}
func (c *Container) GetLogger() (m.Logger, error) {
	if c.logger == nil {
		c.logger = NewAppEngineLogger(c.GetContext())
	}
	return c.logger, nil
}
func (c *Container) MustGetLogger() m.Logger {
	l, _ := c.GetLogger()
	return l
}

// GetContext returns a context
// optionally build the context from the ContextFactory if not nil
func (c *Container) GetContext() context.Context {
	if c.Context == nil {
		if c.ContextFactory != nil {
			c.Context = c.ContextFactory.Create(c.GetRequest())
		} else {
			c.Context = appengine.NewContext(c.GetRequest())
		}
	}
	return c.Context
}

func MustParse(layout string, value string) time.Time {
	t, err := time.Parse(layout, value)
	if err != nil {
		panic(err)
	}
	return t
}

const Rfc2822 = "Mon, 02 Jan 2006 15:04:05 -0700"

// GetMigrations get al list of migrations
func GetMigrations() []*Migration {

	return []*Migration{
		{Name: "001-initial", Created: MustParse("02 Jan 06 15:04 -0700", "18 Oct 16 08:49 +0200"), Task: func(ctx context.Context) error {
			userNickname := "Anonymous"
			count, err := datastore.NewQuery(Kind.Users).Filter("Nickname =", userNickname).Limit(1).Count(ctx)
			if err == nil && count == 0 {
				user := &User{Nickname: userNickname}
				err = NewUserRepository(ctx).Create(user)
			}
			return err
		}}, {
			Name: "002-categories", Created: MustParse(Rfc2822, "Fri, 21 Oct 2016 09:11:26 +0200"), Task: func(ctx context.Context) error {
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
						return err
					}
				}
				return nil
			},
		},
	}
}

// ExecuteMigrations execute all migrations
func ExecuteMigrations(ctx context.Context, migrations []*Migration) error {
	for _, migration := range migrations {
		count, err := datastore.NewQuery(Kind.Migrations).Filter("Name =", migration.Name).Limit(1).Count(ctx)
		if err == nil && count == 0 {
			err = migration.Task(ctx)
			if err == nil {

				migration.ID, _, err = datastore.AllocateIDs(ctx, Kind.Migrations, nil, 1)
				if err == nil {
					_, err = datastore.Put(ctx, datastore.NewKey(ctx, Kind.Migrations, "", migration.ID, nil), migration)
				}
			}
			return err
		}
		if err != nil {
			return err
		}
	}
	return nil
}

type BeforeEntityCreateEvent struct {
	Entity
}

func BeforeEntityCreateListener(e Event) error {
	switch event := e.(type) {
	case BeforeEntityCreateEvent:
		if e, ok := event.Entity.(CreatedUpdatedSetter); ok {
			e.SetCreated(time.Now())
			e.SetUpdated(time.Now())
		}
		if e, ok := event.Entity.(VersionGetterSetter); ok {
			e.SetVersion(1)
		}
	}
	return nil
}

type BeforeEntityUpdateEvent struct {
	Old Entity
	New Entity
}

func BeforeEntityUpdateListener(e Event) error {
	switch event := e.(type) {
	case BeforeEntityUpdateEvent:
		if entity, ok := event.Old.(VersionGetterSetter); ok {
			if old, new := entity, event.New.(VersionGetterSetter); old.GetVersion() != new.GetVersion() {
				return fmt.Errorf("Versions do not match old : %d , new : %d", old.GetVersion(), new.GetVersion())
			} else {
				new.SetVersion(old.GetVersion() + 1)
			}
		}
		if entity, ok := event.New.(CreatedUpdatedSetter); ok {
			entity.SetUpdated(time.Now())
		}
		event.New.SetID(event.Old.GetID())
	}
	return nil
}

type BeforeEntityDeleteEvent struct {
	Entity
}
