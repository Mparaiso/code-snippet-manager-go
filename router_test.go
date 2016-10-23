package smartsnippets_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	app "github.com/Mparaiso/code-snippet-manager-go"
	"github.com/Mparaiso/simple-middleware-go"
)

func ExampleRouter() {
	router := app.NewRouter()
	router.Use(func(next middleware.Handler) middleware.Handler {
		return func(container middleware.Container) {
			container.GetResponseWriter().Header().Add("X-Special", "Yes")
			next(container)
		}
	})
	router.Get("/greetings/:firstname/:lastname", func(container middleware.Container) {
		fmt.Fprintf(container.GetResponseWriter(), "Hello %s %s !",
			container.GetRequest().URL.Query().Get(":firstname"),
			container.GetRequest().URL.Query().Get(":lastname"),
		)
	})
	response := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "http://example.com/greetings/John-Rodger/Doe", nil)
	router.Compile().ServeHTTP(response, request)
	fmt.Println(response.Body.String())
	fmt.Println(response.Header().Get("X-Special"))
	// Output:
	// Hello John-Rodger Doe !
	// Yes
}

func ExampleRouter_Sub() {
	// Router.Sub allows router inheritance
	// SubRouters can be created to allow a custom middleware queue
	// executed by "sub" handlers
	router := app.NewRouter()
	router.Sub("/sub/").
		Use(func(next middleware.Handler) middleware.Handler {
			return func(c middleware.Container) {
				// Will only get executed by by handlers defined
				// in that sub router
				c.GetResponseWriter().Header().Set("X-Sub", "Yes")
			}
		}).
		Get("/", func(c middleware.Container) {
			fmt.Fprint(c.GetResponseWriter(), "Sub")
		})
	response := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "http://example.com/sub/", nil)
	router.Compile().ServeHTTP(response, request)
	fmt.Println(response.Header().Get("X-Sub"))
	// Output:
	// Yes
}

func ExampleRouter_Mount() {
	// Mount allows to define modules
	// That can be reused in different application
	router := app.NewRouter()
	router.Mount("/mount", NewTestModule())
	response := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "http://example.com/mount", nil)
	router.Compile().ServeHTTP(response, request)
	fmt.Println("Status", response.Code)
	// Output:
	// Status 200
}

type ContainerDecorator struct {
	middleware.Container
}

func Decorate(container middleware.Container) middleware.Container {
	return ContainerDecorator{container}
}

type TestModule struct{}

func NewTestModule() *TestModule {
	return &TestModule{}
}

func (module TestModule) Connect(collection *app.RouteCollection) {
	collection.
		Use(func(next middleware.Handler) middleware.Handler {
			return func(container middleware.Container) {
				next(Decorate(container))
			}
		}).
		Get("/", func(c middleware.Container) {
			_, ok := c.(ContainerDecorator)
			if !ok {
				c.Error(fmt.Errorf("Error container is not a ContainerDecorator"), 500)
				return
			}
			c.GetResponseWriter().WriteHeader(200)
		})

}
