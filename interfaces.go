package smartsnippets

import (
	"net/http"
	"time"

	"github.com/Mparaiso/simple-validation-go"

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

// CreatedUpdatedSetter is capable to set created and updated timestamps on a struct
type CreatedUpdatedSetter interface {
	SetCreated(date time.Time)
	SetUpdated(date time.Time)
}
type VersionGetterSetter interface {
	SetVersion(int64)
	GetVersion() int64
}

type UniqueEntityValidatorProvider interface {
	UniqueEntityValidator(field string, values map[string]interface{}, errors validation.Error)
}

type ExistingEntityValidatorProvider interface {
	ExistingEntityValidator(field string, entityName string, values map[string]interface{}, errors validation.Error)
}
