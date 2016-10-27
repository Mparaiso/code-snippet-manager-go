package smartsnippets

import (
	"net/http"

	tiger "github.com/Mparaiso/tiger-go-framework"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
)

// Container is a container
type Container struct {
	tiger.Container
	ContextFactory ContextFactory
	context.Context
	containerOptions ContainerOptions
	logger           tiger.Logger
}

func (c Container) IsDebug() bool {
	return c.containerOptions.Debug
}

// NewContainer creates a new container
func NewContainer(w http.ResponseWriter, r *http.Request) *Container {
	return &Container{Container: tiger.DefaultContainer{Request: r, ResponseWriter: w}}
}
func (c *Container) SetContainerOptions(options ContainerOptions) {
	c.containerOptions = options
}
func (c *Container) Error(err error, statusCode int) {
	if c.IsDebug() {
		c.Container.Error(err, statusCode)
	} else {
		c.Container.Error(tiger.StatusError(statusCode), statusCode)
		c.MustGetLogger().Log(tiger.Error, err)
	}
}
func (c *Container) GetLogger() (tiger.Logger, error) {
	if c.logger == nil {
		c.logger = NewAppEngineLogger(c.GetContext())
	}
	return c.logger, nil
}
func (c *Container) MustGetLogger() tiger.Logger {
	l, _ := c.GetLogger()
	return l
}

type ContainerOptions struct {
	Debug bool
}

// GetContext returns a context
// optionally build the context from the ContextFactory if not nil
func (c *Container) GetContext() context.Context {
	if c.Context == nil {
		if c.ContextFactory != nil {
			c.Context = c.ContextFactory.Create(c.GetRequest())
		} else {
			c.Context = appengine.NewContext(c.GetRequest())
		}
	}
	return c.Context
}
