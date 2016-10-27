package smartsnippets

import (
	"fmt"
	"reflect"

	"github.com/Mparaiso/tiger-go-framework/signal"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"
)

// Kind list app kinds
var Kind = struct{ Users, Migrations, Snippets, Categories, Roles, UserRoles string }{
	"Users", "Migrations", "Snippets", "Categories", "Roles", "UserRoles",
}

// DefaultRepository is the default implementation of Repository
type DefaultRepository struct {
	Context context.Context
	Kind    string
	Signal  signal.Signal
}

// NewDefaultRepository creates a new DefaultRepository
func NewDefaultRepository(ctx context.Context, kind string, listeners ...signal.Listener) *DefaultRepository {
	defaultRepository := &DefaultRepository{Context: ctx, Kind: kind}
	defaultRepository.Signal = signal.NewDefaultSignal()
	defaultRepository.Signal.Add(signal.ListenerFunc(BeforeEntityCreatedListener))
	defaultRepository.Signal.Add(signal.ListenerFunc(BeforeEntityUpdatedListener))
	for _, listener := range listeners {
		defaultRepository.Signal.Add(listener)
	}
	return defaultRepository
}

type ContextValue int

const (
	ParentKey ContextValue = iota
)

var (
	ErrParentKeyNotFound = fmt.Errorf("ErrParentKeyNotFound")
)

func (repository DefaultRepository) GetParentKey() (*datastore.Key, error) {
	//	if parentKey, ok := repository.Context.Value(ParentKey).(*datastore.Key); ok {
	//		return parentKey, nil
	//	}
	//	return nil, ErrParentKeyNotFound
	return GetRootKey(repository.Context), nil
}

// Create an entity
func (repository DefaultRepository) Create(entity Entity) error {
	parentKey, err := repository.GetParentKey()
	if err != nil {
		return err
	}
	low, _, err := datastore.AllocateIDs(repository.Context, repository.Kind, parentKey, 1)
	entity.SetID(low)

	if err == nil {
		if repository.Signal != nil {
			err = repository.Signal.Dispatch(BeforeEntityCreatedEvent{entity})
			if err != nil {
				return err
			}
		}
		key := datastore.NewKey(repository.Context, repository.Kind, "", entity.GetID(), parentKey)
		_, err = datastore.Put(repository.Context, key, entity)
	}
	return err
}

// Update an entity
func (repository DefaultRepository) Update(entity Entity) error {
	parentKey, err := repository.GetParentKey()
	if err != nil {
		return err
	}
	key := datastore.NewKey(repository.Context, repository.Kind, "", entity.GetID(), parentKey)
	old := reflect.New(reflect.Indirect(reflect.ValueOf(entity)).Type()).Interface()
	err = datastore.Get(repository.Context, key, old)
	if err != nil {
		return err
	}
	if repository.Signal != nil {
		err := repository.Signal.Dispatch(BeforeEntityUpdatedEvent{old.(Entity), entity})
		if err != nil {
			return err
		}
	}
	_, err = datastore.Put(repository.Context, key, entity)
	return err
}

// Delete an entity
func (repository DefaultRepository) Delete(entity Entity) error {
	parentKey, err := repository.GetParentKey()
	if err != nil {
		return err
	}
	key := datastore.NewKey(repository.Context, repository.Kind, "", entity.GetID(), parentKey)
	if repository.Signal != nil {
		err = repository.Signal.Dispatch(BeforeEntityDeletedEvent{entity})
		if err != nil {
			return err
		}
	}
	return datastore.Delete(repository.Context, key)
}

// FindByID gets an entity by id
func (repository DefaultRepository) FindByID(id int64, entity Entity) error {
	parentKey, err := repository.GetParentKey()
	if err != nil {
		return err
	}
	key := datastore.NewKey(repository.Context, repository.Kind, "", id, parentKey)
	return datastore.Get(repository.Context, key, entity)
}

// FindAll returns all entities
func (repository DefaultRepository) FindAll(entities interface{}) error {
	parentKey, err := repository.GetParentKey()
	if err != nil {
		return err
	}
	_, err = datastore.NewQuery(repository.Kind).Ancestor(parentKey).GetAll(repository.Context, entities)
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
	parentKey, err := repository.GetParentKey()
	if err != nil {
		return err
	}
	q := repository.createQuery(query)
	_, err = q.Ancestor(parentKey).GetAll(repository.Context, result)
	return err
}

func (repository DefaultRepository) Count(
	query Query) (int, error) {
	parentKey, err := repository.GetParentKey()
	if err != nil {
		return 0, err
	}
	return repository.createQuery(query).Ancestor(parentKey).Count(repository.Context)
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
	*RoleRepository
	*UserRoleRepository
}

func NewUserRepository(ctx context.Context) *UserRepository {

	repository := &UserRepository{Repository: NewDefaultRepository(ctx, Kind.Users)}
	repository.RoleRepository = NewRoleRepository(ctx)
	repository.UserRoleRepository = NewUserRoleRepository(ctx)
	return repository
}

func (u *UserRepository) Create(entity Entity) error {
	if _, ok := entity.(*User); !ok {
		return fmt.Errorf("Entity is ot of type *User")
	}

	err := u.Repository.Create(entity)
	if err != nil {
		return err
	}
	roles := []*Role{}
	err = u.RoleRepository.FindBy(Query{Query: map[string]interface{}{"Name=": "User"}, Limit: 1}, &roles)
	if err != nil {
		return err
	}
	if len(roles) == 0 {
		return fmt.Errorf("Role with Name User not found")
	}
	userRole := &UserRole{UserID: entity.GetID(), RoleID: roles[0].GetID()}
	err = u.UserRoleRepository.Create(userRole)
	return err
}

type RoleRepository struct {
	Repository
}

func NewRoleRepository(ctx context.Context) *RoleRepository {
	return &RoleRepository{NewDefaultRepository(ctx, Kind.Roles)}
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

type UserRoleRepository struct {
	Repository
}

func NewUserRoleRepository(ctx context.Context) *UserRoleRepository {
	return &UserRoleRepository{NewDefaultRepository(ctx, Kind.UserRoles)}
}
