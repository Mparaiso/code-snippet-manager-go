package smartsnippets

import (
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"time"

	"golang.org/x/net/context"

	"sync"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"

	m "github.com/Mparaiso/simple-middleware-go"
	matcher "github.com/Mparaiso/simple-router-go"
)

// An Entity is a datastore entity
//
// GetID returns the ID
//
// SetID sets the ID
type Entity interface {
	GetID() int64
	SetID(int64)
}

// Signal is an implementation of the signal pattern
type Signal interface {
	Add(Listener)
	Remove(Listener)
	Dispatch(data interface{}) error
}

// Listener handle signals
type Listener interface {
	Handle(data interface{}) error
}

type kind struct{ Users, Migrations, Snippets string }

// Kind list app kinds
var Kind = kind{"Users", "Migrations", "Snippets"}

func init() {
	app := new(App)
	http.Handle("/", app)
}

var _ matcher.MatcherProvider = new(Route)

// Route is an app route
type Route struct {
	Handler  m.Handler
	Matchers []matcher.Matcher
}

// GetMatchers return the request matchers
func (route *Route) GetMatchers() matcher.Matchers {
	return route.Matchers
}

func (route Route) String() string {
	representation := "("
	for _, matcher := range route.Matchers {
		representation += fmt.Sprint(matcher, ",")
	}
	return representation[:len(representation)-1] + ")"
}

// App is the web application
type App struct {
	sync.Once
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	container := NewContainer(w, r)
	a.Do(func() {
		ctx := container.GetContext()
		if err := ExecuteMigrations(ctx, GetMigrations()); err != nil {
			log.Errorf(ctx, "Error during migration '%s'.", err.Error())
		} else {
			log.Infof(ctx, "Migrations done")
		}
	})
	snippetEndpoint := &EndPoint{EntityType: reflect.TypeOf(&Snippet{}), Kind: Kind.Snippets}
	requestMatcher := matcher.NewRequestMatcher(matcher.MatcherProviders{
		&Route{warmup, matcher.Matchers{matcher.NewURLMatcher("/_ah/warmup")}},
		&Route{snippetEndpoint.Index,
			matcher.Matchers{matcher.NewMethodMatcher("GET"),
				matcher.NewRegexMatcher(regexp.MustCompile(`^/snippets/?$`))}},
		&Route{snippetEndpoint.Post,
			matcher.Matchers{matcher.NewMethodMatcher("POST"),
				matcher.NewRegexMatcher(regexp.MustCompile(`^/snippets/?$`))}},
		&Route{snippetEndpoint.Get,
			matcher.Matchers{matcher.NewMethodMatcher("GET"),
				matcher.NewRegexMatcher(regexp.MustCompile(`^/snippets/(?P<id>\d+)/?$`))}},
		&Route{index, matcher.Matchers{matcher.NewMethodMatcher("GET"), matcher.NewURLMatcher("/")}},
		&Route{name, matcher.Matchers{matcher.NewRegexMatcher(regexp.MustCompile(`^/(?P<name>[\w \s]+)/?$`))}},
	})

	if matcher := requestMatcher.Match(r); matcher != nil {
		matcher.(*Route).Handler(container)
	} else {
		http.NotFound(w, r)
	}
}

func onStart(c m.Container) {
	ctx := appengine.NewContext(c.Request())
	log.Infof(ctx, "=> start up")
}

func index(c m.Container) {
	fmt.Fprint(c.ResponseWriter(), "Hello Smart Snippets")
}

func name(c m.Container) {
	name := c.Request().URL.Query().Get("@name")
	fmt.Fprintf(c.ResponseWriter(), "Hello %s", name)
}

func warmup(c m.Container) {
	ctx := appengine.NewContext(c.Request())
	log.Infof(ctx, "=> warmup done")
}

// ContextProvider provides a context
type ContextProvider interface {
	GetContext() context.Context
}

// Container is a container
type Container struct {
	m.Container
	context.Context
}

// NewContainer creates a new container
func NewContainer(w http.ResponseWriter, r *http.Request) *Container {
	return &Container{Container: m.DefaultContainer{Req: r, RW: w}}
}

// GetContext returns a context
func (c *Container) GetContext() context.Context {
	if c.Context == nil {
		c.Context = appengine.NewContext(c.Request())
	}
	return c.Context
}

// Migration is a DB migration
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
	ID       int64
	Nickname string
	Created  time.Time
	Updated  time.Time
	Version  int64
}

func (u User) GetID() int64               { return u.ID }
func (u *User) SetID(id int64)            { u.ID = id }
func (u User) GetVersion() int64          { return u.Version }
func (u *User) SetVersion(version int64)  { u.Version = version }
func (u *User) SetCreated(date time.Time) { u.Created = date }
func (u *User) SetUpdated(date time.Time) { u.Updated = date }

type CreatedUpdatedSetter interface {
	SetCreated(date time.Time)
	SetUpdated(date time.Time)
}
type VersionGetterSetter interface {
	SetVersion(int64)
	GetVersion() int64
}

type UserRepository struct {
	Repository
}

func NewUserRepository(ctx context.Context) *UserRepository {
	return &UserRepository{NewDefaultRepository(ctx, Kind.Users)}
}

type MigrationRepository struct {
	Repository
}

func NewMigrationRepository(ctx context.Context) *MigrationRepository {
	return &MigrationRepository{NewDefaultRepository(ctx, Kind.Migrations)}
}

type BeforeCreateListener interface {
	BeforeCreate(ctx context.Context) error
}

type BeforeUpdateListener interface {
	BeforeUpdate(ctx context.Context) error
}

// Repository is a entity repository
type Repository interface {
	Create(entity Entity) error
	Update(entity Entity) error
	Delete(entity Entity) error
	FindByID(id int64, entity Entity) error
	FindAll(entities interface{}) error
}

// DefaultRepository is the default implementation of Repository
type DefaultRepository struct {
	Context            context.Context
	Kind               string
	BeforeCreateSignal Signal
	BeforeUpdateSignal Signal
	BeforeDeleteSignal Signal
}

func NewDefaultRepository(ctx context.Context, kind string) *DefaultRepository {
	defaultRepository := &DefaultRepository{Context: ctx, Kind: kind}
	defaultRepository.BeforeCreateSignal = NewDefaultSignal()
	defaultRepository.BeforeCreateSignal.Add(ListenerFunc(BeforeCreateListenerFunc))
	defaultRepository.BeforeUpdateSignal = NewDefaultSignal()
	defaultRepository.BeforeUpdateSignal.Add(ListenerFunc(BeforeUpdateListenerFunc))
	return defaultRepository
}

// Create an entity
func (repository DefaultRepository) Create(entity Entity) error {
	low, _, err := datastore.AllocateIDs(repository.Context, repository.Kind, nil, 1)
	entity.SetID(low)

	if err == nil {
		if repository.BeforeCreateSignal != nil {
			err = repository.BeforeCreateSignal.Dispatch(entity)
			if err != nil {
				return err
			}
		}
		key := datastore.NewKey(repository.Context, repository.Kind, "", entity.GetID(), nil)
		_, err = datastore.Put(repository.Context, key, entity)
	}
	return err
}

// Update an entity
func (repository DefaultRepository) Update(entity Entity) error {
	key := datastore.NewKey(repository.Context, repository.Kind, "", entity.GetID(), nil)
	if repository.BeforeUpdateSignal != nil {
		err := repository.BeforeUpdateSignal.Dispatch(entity)
		if err != nil {
			return err
		}
	}
	_, err := datastore.Put(repository.Context, key, entity)
	return err
}

// Delete an entity
func (repository DefaultRepository) Delete(entity Entity) error {
	key := datastore.NewKey(repository.Context, repository.Kind, "", entity.GetID(), nil)
	if repository.BeforeDeleteSignal != nil {
		err := repository.BeforeDeleteSignal.Dispatch(entity)
		if err != nil {
			return err
		}
	}
	return datastore.Delete(repository.Context, key)
}

// FindByID gets an entity by id
func (repository DefaultRepository) FindByID(id int64, entity Entity) error {
	key := datastore.NewKey(repository.Context, repository.Kind, "", id, nil)
	return datastore.Get(repository.Context, key, entity)
}

func (repository DefaultRepository) FindAll(entities interface{}) error {
	_, err := datastore.NewQuery(repository.Kind).GetAll(repository.Context, entities)
	return err
}

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
}

func MustParse(layout string, value string) time.Time {
	t, err := time.Parse(layout, value)
	if err != nil {
		panic(err)
	}
	return t
}

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
		}},
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
		}
		if err != nil {
			return err
		}
	}
	return nil
}

type ListenerFunc func(data interface{}) error

func (l ListenerFunc) Handle(data interface{}) error {
	return l(data)
}

type DefaultSignal struct {
	Listeners []Listener
}

func NewDefaultSignal() *DefaultSignal {
	return &DefaultSignal{Listeners: []Listener{}}
}

func (signal *DefaultSignal) Add(l Listener) {
	if signal.IndexOf(l) != -1 {
		return
	}
	signal.Listeners = append(signal.Listeners, l)
}

func (signal *DefaultSignal) IndexOf(l Listener) int {
	for i, listener := range signal.Listeners {
		if listener == l {
			return i
		}
	}
	return -1
}

func (signal *DefaultSignal) Remove(l Listener) {
	index := signal.IndexOf(l)
	if index == -1 {
		return
	}
	head := signal.Listeners[:index]
	if index == len(signal.Listeners)-1 {
		signal.Listeners = head
	} else {
		signal.Listeners = append(head, signal.Listeners[index+1:]...)
	}
}

func (signal *DefaultSignal) Dispatch(data interface{}) error {
	for _, listener := range signal.Listeners {
		if err := listener.Handle(data); err != nil {
			return err
		}
	}
	return nil
}

func BeforeCreateListenerFunc(entity interface{}) error {
	if e, ok := entity.(CreatedUpdatedSetter); ok {
		e.SetCreated(time.Now())
		e.SetUpdated(time.Now())
	}
	if e, ok := entity.(VersionGetterSetter); ok {
		e.SetVersion(1)
	}
	return nil
}

func BeforeUpdateListenerFunc(entity interface{}) error {
	if e, ok := entity.(CreatedUpdatedSetter); ok {
		e.SetUpdated(time.Now())
	}
	if e, ok := entity.(VersionGetterSetter); ok {
		e.SetVersion(e.GetVersion() + 1)
	}
	return nil
}
