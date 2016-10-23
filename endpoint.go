package smartsnippets

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"reflect"

	m "github.com/Mparaiso/simple-middleware-go"
)

type RepositoryProvider interface {
	GetRepository() (Repository, error)
}

type BeforeCreateEvent struct{}
type AfterCreateEvent struct{}

type BeforeUpdateEvent struct{}
type AfterUpdateEvent struct{}

type BeforeDeleteEvent struct{}
type AfterDeleteEvent struct{}

type EndPointContainer interface {
	m.Container
	GetRepository() Repository
	GetPrototype() reflect.Type
}

type EndPointContainerFactory interface {
	Create(m.Container) EndPointContainer
}

// EndPoint is a rest endpoint
type EndPoint struct {
	Prototype                 interface{}
	cached                    bool
	protoType                 reflect.Type
	RepositoryProviderFactory func(m.Container) RepositoryProvider
	IndexHandler              func(m.Container)
	Signal                    Signal
	EndPointContainerFactory
}

func NewEndpoint(endpointContainerFactory EndPointContainerFactory) *EndPoint {
	return &EndPoint{EndPointContainerFactory: endpointContainerFactory}
}

// Handle handles HTTP requests
func (e EndPoint) Connect(routeCollection *RouteCollection) {
	routeCollection.
		Use(func(next m.Handler) m.Handler {
			return func(c m.Container) {
				c.(*Container).MustGetLogger().Log(Debug, "Setting endpoint container")
				next(e.EndPointContainerFactory.Create(c))
			}
		}).
		Get("/", e.Index).
		Post("/", e.Post).
		Put("/:id", e.Put).
		Delete("/:id", e.Delete).
		Get("/:id", e.Get)
}

// Index list resources
func (e EndPoint) Index(c m.Container) {
	if e.IndexHandler != nil {
		e.IndexHandler(c)
	}
	container := c.(EndPointContainer)
	repository := container.GetRepository()
	entities := reflect.New(reflect.SliceOf(container.GetPrototype())).Interface()
	err := repository.FindAll(entities)
	if err != nil {
		container.Error(err, http.StatusInternalServerError)
		return
	}
	json.NewEncoder(container.GetResponseWriter()).Encode(entities)
}

// Get fetches a resource
func (e EndPoint) Get(c m.Container) {
	container := c.(EndPointContainer)
	var id int64
	_, err := fmt.Sscanf(container.GetRequest().URL.Query().Get(":id"), "%d", &id)
	if err != nil {
		container.Error(err, http.StatusBadRequest)
		return
	}
	entity := reflect.New(container.GetPrototype()).Interface()
	repository := container.GetRepository()
	err = repository.FindByID(id, entity.(Entity))
	if err != nil {
		container.Error(err, http.StatusInternalServerError)
		return
	}
	json.NewEncoder(container.GetResponseWriter()).Encode(entity)
}

// Put updates a resource
func (e EndPoint) Put(c m.Container) {
	container := c.(EndPointContainer)
	entity := reflect.New(container.GetPrototype()).Interface()
	var id int64
	_, err := fmt.Sscanf(container.GetRequest().URL.Query().Get(":id"), "%d", &id)
	if err != nil {
		container.Error(err, http.StatusBadRequest)
		return
	}
	repository := container.GetRepository()
	err = repository.FindByID(id, entity.(Entity))
	if err != nil {
		container.Error(err, http.StatusNotFound)
		return
	}
	candidate := reflect.New(container.GetPrototype()).Interface()
	json.NewDecoder(container.GetRequest().Body).Decode(candidate)
	candidate.(Entity).SetID(id)
	err = e.Signal.Dispatch(&BeforeEntityUpdateEvent{})
	if err != nil {
		container.Error(err, http.StatusInternalServerError)
		return
	}
	err = repository.Update(candidate.(Entity))
	if err != nil {
		container.Error(err, http.StatusInternalServerError)
		return
	}

	if err = e.Signal.Dispatch(&AfterUpdateEvent{}); err != nil {
		container.Error(err, http.StatusInternalServerError)
		return
	}
	container.GetResponseWriter().WriteHeader(http.StatusOK)
}

// Delete deletes a resource
func (e EndPoint) Delete(c m.Container) {
	container := c.(EndPointContainer)
	entity := reflect.New(container.GetPrototype()).Interface()
	var id int64
	_, err := fmt.Sscanf(container.GetRequest().URL.Query().Get(":id"), "%d", &id)
	if err != nil {
		container.Error(err, http.StatusBadRequest)
		return
	}
	repository, err := e.RepositoryProviderFactory(container).GetRepository()
	if err != nil {
		container.Error(err, http.StatusInternalServerError)
		return
	}
	err = repository.FindByID(id, entity.(Entity))
	if err != nil {
		container.Error(err, http.StatusNotFound)
		return
	}
	err = e.Signal.Dispatch(&BeforeDeleteEvent{})
	if err != nil {
		container.Error(err, http.StatusInternalServerError)
		return
	}
	err = repository.Delete(entity.(Entity))
	if err != nil {
		container.Error(err, http.StatusInternalServerError)
		return
	}
	err = e.Signal.Dispatch(&AfterDeleteEvent{})
	if err != nil {
		container.Error(err, http.StatusInternalServerError)
		return
	}
	container.GetResponseWriter().WriteHeader(http.StatusOK)
}

// Post creates a resource
func (e EndPoint) Post(c m.Container) {
	container := c.(EndPointContainer)
	entity := reflect.New(container.GetPrototype()).Interface()

	decoder := json.NewDecoder(container.GetRequest().Body)
	err := decoder.Decode(entity)
	if err != nil {
		container.Error(err, http.StatusBadRequest)
		return
	}
	repository, err := e.RepositoryProviderFactory(container).GetRepository()
	if err != nil {
		container.Error(err, http.StatusInternalServerError)
		return
	}
	err = repository.Create(entity.(Entity))
	if err != nil {
		container.Error(err, http.StatusInternalServerError)
		return
	}
	location := path.Join(container.GetRequest().URL.Path, fmt.Sprintf("%d", entity.(Entity).GetID()))
	container.GetRequest().Method = "GET"
	http.Redirect(container.GetResponseWriter(), container.GetRequest(), location, 303)
}
