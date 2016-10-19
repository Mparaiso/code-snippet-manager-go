package smartsnippets

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"reflect"
	"strconv"

	m "github.com/Mparaiso/simple-middleware-go"
	"google.golang.org/appengine/log"
)

// EndPoint is a rest endpoint
type EndPoint struct {
	Kind       string
	EntityType reflect.Type
}

// Handle handles HTTP requests
/*func (e EndPoint) Handle(container m.Container) {
	requestMatcher := matcher.NewRequestMatcher(matcher.MatcherProviders{
		&Route{e.Get, matcher.Matchers{matcher.NewRegexMatcher(regexp.MustCompile(`/(?P<id>\d+)/?$`)), matcher.NewMethodMatcher("GET")}},
		&Route{e.Put, matcher.Matchers{matcher.NewRegexMatcher(regexp.MustCompile(`/(?P<id>\d+)/?$`)), matcher.NewMethodMatcher("PUT")}},
		&Route{e.Delete, matcher.Matchers{matcher.NewRegexMatcher(regexp.MustCompile(`/(?P<id>\d+)/?$`)), matcher.NewMethodMatcher("DELETE")}},
		&Route{e.Index, matcher.Matchers{matcher.NewRegexMatcher(regexp.MustCompile(`/$`)), matcher.NewMethodMatcher("GET")}},
		&Route{e.Post, matcher.Matchers{matcher.NewRegexMatcher(regexp.MustCompile(`/$`)), matcher.NewMethodMatcher("POST")}},
	})
	if match := requestMatcher.Match(container.Request()); match != nil {
		match.(*Route).Handler(container)
	} else {
		http.NotFound(container.ResponseWriter(), container.Request())
	}
}*/

// Index list resources
func (e EndPoint) Index(c m.Container) {
	container := c.(interface {
		ContextProvider
		m.Container
	})
	repository := NewDefaultRepository(container.GetContext(), e.Kind)
	//entities := reflect.MakeSlice(reflect.SliceOf(e.EntityType), 0, 0).Interface()
	entities := reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(&Snippet{})), 0, 0).Interface()
	err := repository.FindAll(entities)
	if err != nil {
		http.Error(c.ResponseWriter(), http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Errorf(container.GetContext(), "repository.FindAll %s", err)
		return
	}
	encoder := json.NewEncoder(c.ResponseWriter())
	encoder.Encode(entities)

}

// Get fetches a resource
func (e EndPoint) Get(c m.Container) {
	container := c.(interface {
		ContextProvider
		m.Container
	})
	id := container.Request().URL.Query().Get("@id")
	ID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		log.Errorf(container.GetContext(), "Error converting ID %s", err)
		http.Error(container.ResponseWriter(), err.Error(), http.StatusInternalServerError)
		return
	}
	entity := reflect.New(e.EntityType.Elem()).Interface()
	err = NewDefaultRepository(container.GetContext(), e.Kind).FindByID(ID, entity.(Entity))
	if err != nil {
		log.Errorf(container.GetContext(), "Error finding entity with ID  %d", ID)
		http.Error(container.ResponseWriter(), err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(container.ResponseWriter()).Encode(entity)
}

// Put updates a resource
func (EndPoint) Put(container m.Container) {}

// Delete deletes a resource
func (EndPoint) Delete(container m.Container) {}

// Post creates a resource
func (e EndPoint) Post(c m.Container) {
	container := c.(interface {
		ContextProvider
		m.Container
	})
	entity := reflect.New(e.EntityType.Elem()).Interface()

	decoder := json.NewDecoder(container.Request().Body)
	err := decoder.Decode(entity)
	if err != nil {
		http.Error(container.ResponseWriter(), http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		log.Errorf(container.GetContext(), "%s", err)
		return
	}
	repository := NewDefaultRepository(container.GetContext(), e.Kind)
	err = repository.Create(entity.(Entity))
	if err != nil {
		http.Error(container.ResponseWriter(), http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Errorf(container.GetContext(), "%s", err)
		return
	}
	location := path.Join(container.Request().URL.Path, fmt.Sprintf("%d", entity.(Entity).GetID()))
	container.Request().Method = "GET"
	http.Redirect(container.ResponseWriter(), container.Request(), location, 303)
}
