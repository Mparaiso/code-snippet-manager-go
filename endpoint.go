package smartsnippets

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"reflect"

	tiger "github.com/Mparaiso/tiger-go-framework"
	"github.com/Mparaiso/tiger-go-framework/signal"
)

type EndPointOptions struct {
	Commands map[string]bool
}

// EndPoint is a rest endpoint
type EndPoint struct {
	Prototype                 interface{}
	cached                    bool
	protoType                 reflect.Type
	RepositoryProviderFactory func(tiger.Container) RepositoryProvider
	IndexHandler              func(tiger.Container)
	Signal                    signal.Signal
	EndPointContainerFactory
	Options EndPointOptions
}

func NewEndpoint(endpointContainerFactory EndPointContainerFactory, endpointOptions ...EndPointOptions) *EndPoint {

	endpoint := &EndPoint{EndPointContainerFactory: endpointContainerFactory}
	if len(endpointOptions) > 0 {
		endpoint.Options = endpointOptions[0]
	}
	return endpoint
}

// Connect handles HTTP requests
func (e EndPoint) Connect(routeCollection *tiger.RouteCollection) {

	routeCollection.
		Use(func(c tiger.Container, next tiger.Handler) {
			next(e.EndPointContainerFactory.Create(c))
		})
	if len(e.Options.Commands) == 0 || e.Options.Commands["INDEX"] {
		routeCollection.Get("/", e.Wrap(e.Index))
	}
	if len(e.Options.Commands) == 0 || e.Options.Commands["POST"] {
		routeCollection.Post("/", e.Wrap(e.Post))
	}
	if len(e.Options.Commands) == 0 || e.Options.Commands["PUT"] {
		routeCollection.Put("/:id", e.Wrap(e.Put))
	}
	if len(e.Options.Commands) == 0 || e.Options.Commands["DELETE"] {
		routeCollection.Delete("/:id", e.Wrap(e.Delete))
	}
	if len(e.Options.Commands) == 0 || e.Options.Commands["GET"] {
		routeCollection.Get("/:id", e.Wrap(e.Get))
	}
}

func (EndPoint) Wrap(handler func(c EndPointContainer)) func(c tiger.Container) {
	return func(c tiger.Container) {
		handler(c.(EndPointContainer))
	}
}

// Index list resources
func (e EndPoint) Index(container EndPointContainer) {
	if e.IndexHandler != nil {
		e.IndexHandler(container)
		return
	}
	repository := container.GetRepository()
	entities := reflect.New(reflect.SliceOf(container.GetPrototype())).Interface()
	err := repository.FindAll(entities)
	if err != nil {
		container.Error(err, http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(container.GetResponseWriter()).Encode(entities)
	if err != nil {
		container.Error(err, http.StatusInternalServerError)
	}
}

// Get fetches a resource
func (e EndPoint) Get(container EndPointContainer) {

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
	err = json.NewEncoder(container.GetResponseWriter()).Encode(entity)
	if err != nil {
		container.Error(err, http.StatusInternalServerError)
	}
}

// Put updates a resource
func (e EndPoint) Put(container EndPointContainer) {
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
	err = container.GetSignal().Dispatch(&BeforeEntityUpdatedEvent{})
	if err != nil {
		container.Error(err, http.StatusInternalServerError)
		return
	}
	err = repository.Update(candidate.(Entity))
	if err != nil {
		container.Error(err, http.StatusInternalServerError)
		return
	}

	if err = container.GetSignal().Dispatch(&AfterResourceUpdateEvent{}); err != nil {
		container.Error(err, http.StatusInternalServerError)
		return
	}
	container.GetResponseWriter().WriteHeader(http.StatusOK)
}

// Delete deletes a resource
func (e EndPoint) Delete(container EndPointContainer) {

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
	err = container.GetSignal().Dispatch(&BeforeResourceDeleteEvent{})
	if err != nil {
		container.Error(err, http.StatusInternalServerError)
		return
	}
	err = repository.Delete(entity.(Entity))
	if err != nil {
		container.Error(err, http.StatusInternalServerError)
		return
	}
	err = container.GetSignal().Dispatch(&AfterResourceDeleteEvent{})
	if err != nil {
		container.Error(err, http.StatusInternalServerError)
		return
	}
	container.GetResponseWriter().WriteHeader(http.StatusOK)
}

// Post creates a resource
func (e EndPoint) Post(container EndPointContainer) {
	entity := reflect.New(container.GetPrototype()).Interface()

	decoder := json.NewDecoder(container.GetRequest().Body)
	err := decoder.Decode(entity)
	if err != nil {
		container.Error(err, http.StatusBadRequest)
		return
	}
	repository := container.GetRepository()

	err = repository.Create(entity.(Entity))
	if err != nil {
		container.Error(err, http.StatusInternalServerError)
		return
	}
	location := path.Join(container.GetRequest().URL.Path, fmt.Sprintf("%d", entity.(Entity).GetID()))
	container.GetRequest().Method = "GET"
	http.Redirect(container.GetResponseWriter(), container.GetRequest(), location, 303)
}
