package smartsnippets

import (
	"net/http"
	"reflect"
	"time"

	tiger "github.com/Mparaiso/tiger-go-framework"
	"github.com/Mparaiso/tiger-go-framework/signal"
	"github.com/Mparaiso/tiger-go-framework/validator"
	"golang.org/x/net/context"
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

// CreatedUpdatedSetter is capable to set created and updated timestamps on a struct
type CreatedUpdatedEntity interface {
	SetCreated(date time.Time)
	SetUpdated(date time.Time)
}
type VersionedEntity interface {
	SetVersion(int64)
	GetVersion() int64
}

type LockedEntity interface {
	IsLocked() bool
}

// Repository is a entity repository
type Repository interface {
	Create(entity Entity) error
	Update(entity Entity) error
	Delete(entity Entity) error
	FindByID(id int64, entity Entity) error
	FindAll(entities interface{}) error
	FindBy(query Query, result interface{}) error
	Count(query Query) (int, error)
}

// ContextFactory creates a ContextProvider
type ContextFactory interface {
	Create(r *http.Request) context.Context
}

// ContextProvider provides a context
type ContextProvider interface {
	GetContext() context.Context
}
type RepositoryProvider interface {
	GetRepository() Repository
}

type SignalProvider interface {
	GetSignal() signal.Signal
}

type UniqueEntityValidatorProvider interface {
	UniqueEntityValidator(field string, values map[string]interface{}, errors validator.Error)
}

type ExistingEntityValidatorProvider interface {
	ExistingEntityValidator(field string, entityName string, values map[string]interface{}, errors validator.Error)
}

type ContextAwareContainer interface {
	tiger.Container
	ContextProvider
}

type EndPointContainer interface {
	tiger.Container
	RepositoryProvider
	GetPrototype() reflect.Type
	SignalProvider
}

type EndPointContainerFactory interface {
	Create(tiger.Container) EndPointContainer
}
