package smartsnippets

import (
	"reflect"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"
)

// DefaultRepository is the default implementation of Repository
type DefaultRepository struct {
	Context context.Context
	Kind    string
	Signal  Signal
}

// NewDefaultRepository creates a new DefaultRepository
func NewDefaultRepository(ctx context.Context, kind string, listeners ...Listener) *DefaultRepository {
	defaultRepository := &DefaultRepository{Context: ctx, Kind: kind}
	defaultRepository.Signal = NewDefaultSignal()
	defaultRepository.Signal.Add(ListenerFunc(BeforeEntityCreateListener))
	defaultRepository.Signal.Add(ListenerFunc(BeforeEntityUpdateListener))
	for _, listener := range listeners {
		defaultRepository.Signal.Add(listener)
	}
	return defaultRepository
}

// Create an entity
func (repository DefaultRepository) Create(entity Entity) error {
	low, _, err := datastore.AllocateIDs(repository.Context, repository.Kind, nil, 1)
	entity.SetID(low)

	if err == nil {
		if repository.Signal != nil {
			err = repository.Signal.Dispatch(BeforeEntityCreateEvent{entity})
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
	old := reflect.New(reflect.Indirect(reflect.ValueOf(entity)).Type()).Interface()
	err := datastore.Get(repository.Context, key, old)
	if err != nil {
		return err
	}
	if repository.Signal != nil {
		err := repository.Signal.Dispatch(BeforeEntityUpdateEvent{old.(Entity), entity})
		if err != nil {
			return err
		}
	}
	_, err = datastore.Put(repository.Context, key, entity)
	return err
}

// Delete an entity
func (repository DefaultRepository) Delete(entity Entity) error {
	key := datastore.NewKey(repository.Context, repository.Kind, "", entity.GetID(), nil)
	if repository.Signal != nil {
		err := repository.Signal.Dispatch(BeforeEntityDeleteEvent{entity})
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

// FindAll returns all entities
func (repository DefaultRepository) FindAll(entities interface{}) error {
	_, err := datastore.NewQuery(repository.Kind).GetAll(repository.Context, entities)
	return err
}

type Query struct {
	Query  map[string]interface{}
	Order  []string
	Fields []string
	Limit  int
	Offset int
}

func (repository DefaultRepository) FindBy(
	query Query,
	result interface{}) error {
	q := repository.createQuery(query)
	_, err := q.GetAll(repository.Context, result)
	return err
}

func (repository DefaultRepository) Count(
	query Query) (int, error) {
	return repository.createQuery(query).Count(repository.Context)
}

func (repository DefaultRepository) createQuery(query Query) *datastore.Query {
	q := datastore.NewQuery(repository.Kind)
	for key, value := range query.Query {
		q = q.Filter(key, value)
	}
	for _, o := range query.Order {
		q = q.Order(o)
	}
	if len(query.Fields) > 0 {
		q = q.Project(query.Fields...)
	}
	if query.Limit > 0 {
		q = q.Limit(query.Limit)
	}
	return q.Offset(query.Offset)

}

type UserRepository struct {
	Repository
}

func NewUserRepository(ctx context.Context) *UserRepository {
	return &UserRepository{NewDefaultRepository(ctx, Kind.Users)}
}

type CategoryRepository struct {
	Repository
}

func NewCategoryRepository(ctx context.Context) *CategoryRepository {
	return &CategoryRepository{NewDefaultRepository(ctx, Kind.Categories)}
}

type MigrationRepository struct {
	Repository
}

func NewMigrationRepository(ctx context.Context) *MigrationRepository {
	return &MigrationRepository{NewDefaultRepository(ctx, Kind.Migrations)}
}
