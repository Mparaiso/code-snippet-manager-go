package smartsnippets

import (
	"fmt"
	"net/http"
	"reflect"

	"sync"

	"google.golang.org/appengine/log"

	tiger "github.com/Mparaiso/tiger-go-framework"
	"github.com/Mparaiso/tiger-go-framework/signal"
)

// App is the web application
type App struct {
	*sync.Once
	Debug bool
	*tiger.Router
}

func (a App) GetContainer(w http.ResponseWriter, r *http.Request) tiger.Container {
	return NewContainer(w, r)
}

func NewApp() *App {
	app := new(App)
	app.Debug = true
	app.Router = tiger.NewRouter()
	app.Once = new(sync.Once)

	app.ContainerFactory = app

	snippetEndpoint := NewEndpoint(new(SnippetEndPointContainerFactory))
	categoryEndpoint := NewEndpoint(new(CategoryEndPointContainerFactory))
	userEndpoint := NewEndpoint(new(UserEndPointContainerFactory))
	migrationEndpoint := NewEndpoint(new(MigrationEndPointContainerFactory), EndPointOptions{Commands: map[string]bool{"INDEX": true}})
	usersModule := NewUserEndpoint(NewUserEndpointContainerFactory())
	app.Use(func(c tiger.Container, next tiger.Handler) {
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
	}).
		Get("/", index).
		Mount("/users/", usersModule).
		Mount("/snippets", snippetEndpoint).
		Mount("/categories", categoryEndpoint).
		Mount("/users", userEndpoint).
		Mount("/migrations", migrationEndpoint)
	return app
}

type MigrationEndPointContainerFactory struct{}

func (MigrationEndPointContainerFactory) Create(container tiger.Container) EndPointContainer {
	return NewDefaultEndPointContainer(
		Kind.Migrations,
		reflect.TypeOf(Migration{}),
		container.(*Container),
	)
}

type UserEndPointContainerFactory struct {
}

func (UserEndPointContainerFactory) Create(container tiger.Container) EndPointContainer {
	return NewDefaultEndPointContainer(
		Kind.Users,
		reflect.TypeOf(User{}),
		container.(*Container),
	)
}

type SnippetEndPointContainerFactory struct{}

func (SnippetEndPointContainerFactory) Create(container tiger.Container) EndPointContainer {
	return NewDefaultEndPointContainer(
		Kind.Snippets,
		reflect.TypeOf(Snippet{}),
		container.(*Container),
	)
}

type CategoryEndPointContainerFactory struct{}

func (CategoryEndPointContainerFactory) Create(container tiger.Container) EndPointContainer {
	return NewDefaultEndPointContainer(
		Kind.Categories,
		reflect.TypeOf(Category{}),
		container.(*Container),
	)
}

type DefaultEndPointContainer struct {
	prototype reflect.Type
	ContextAwareContainer
	RepositoryProvider
	signal signal.Signal
}

func NewDefaultEndPointContainer(
	kind string,
	prototype reflect.Type,
	container ContextAwareContainer,
	listeners ...signal.Listener,
) *DefaultEndPointContainer {
	enpointContainer := &DefaultEndPointContainer{ContextAwareContainer: container, prototype: prototype}
	enpointContainer.RepositoryProvider = NewAppengineRepositoryProvider(enpointContainer.ContextAwareContainer, kind, listeners...)
	return enpointContainer
}

func (endPointContainer *DefaultEndPointContainer) GetSignal() signal.Signal {
	if endPointContainer.signal == nil {
		endPointContainer.signal = signal.NewDefaultSignal()
	}
	return endPointContainer.signal
}

func (endpointContainer DefaultEndPointContainer) GetPrototype() reflect.Type {
	return endpointContainer.prototype
}

type AppengineRepositoryProvider struct {
	Kind string
	ContextProvider
	Repository
	listeners []signal.Listener
}

func NewAppengineRepositoryProvider(contextProvider ContextProvider, kind string, listeners ...signal.Listener) *AppengineRepositoryProvider {
	return &AppengineRepositoryProvider{Kind: kind, ContextProvider: contextProvider, listeners: listeners}
}

func (provider *AppengineRepositoryProvider) GetRepository() Repository {
	if provider.Repository == nil {
		provider.Repository = NewDefaultRepository(provider.GetContext(), provider.Kind, provider.listeners...)
	}
	return provider.Repository
}

func index(c tiger.Container) {
	fmt.Fprint(c.GetResponseWriter(), "Hello Smart Snippets")
}

func init() {
	http.Handle("/", NewApp().Compile())
}
