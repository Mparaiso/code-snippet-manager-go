package smartsnippets

import (
	"encoding/json"
	"fmt"
	"net/http"

	tiger "github.com/Mparaiso/tiger-go-framework"
	"golang.org/x/crypto/bcrypt"
)

func EncryptPassword(password string) (encryptedPassword string, err error) {
	Bytes, e := bcrypt.GenerateFromPassword([]byte(password), 0)
	if e != nil {
		return "", e
	}
	return string(Bytes), nil
}

type UserEndpointContainerFactory struct{}

func NewUserEndpointContainerFactory() *UserEndpointContainerFactory {
	return &UserEndpointContainerFactory{}
}
func (factory UserEndpointContainerFactory) Create(container tiger.Container) UserEndpointContainer {
	return &DefaultUserEndpointContainer{
		ContextAwareContainer: container.(*Container),
	}
}

type DefaultUserEndpointContainer struct {
	ContextAwareContainer
	UserRepository Repository
}

func (container *DefaultUserEndpointContainer) GetRepository() Repository {
	if container.UserRepository == nil {
		container.UserRepository = NewUserRepository(container.GetContext())
	}
	return container.UserRepository
}

func (container DefaultUserEndpointContainer) Validate(user *User) error {
	validator := &UserValidator{&DefaultUniqueEntityValidatorProvider{container.GetRepository()}}
	return validator.Validate(user)
}

type UserEndpointContainer interface {
	RepositoryProvider
	tiger.Container
	Validate(*User) error
}

type Validator interface {
	Validate(*User) error
}

type UserEndpoint struct {
	*UserEndpointContainerFactory
}

func NewUserEndpoint(containerFactory *UserEndpointContainerFactory) *UserEndpoint {
	return &UserEndpoint{containerFactory}
}
func (module UserEndpoint) Connect(routeCollection *tiger.RouteCollection) {
	routeCollection.
		Use(func(container tiger.Container, next tiger.Handler) {
			next(module.UserEndpointContainerFactory.Create(container))
		}).
		Post("/register", module.Wrap(module.Register))
}
func (module UserEndpoint) Register(container UserEndpointContainer) {
	user := &User{}

	err := json.NewDecoder(container.GetRequest().Body).Decode(user)
	if err != nil {
		container.Error(err, http.StatusBadRequest)
		return
	}

	if err = container.Validate(user); err != nil {
		container.Error(err, http.StatusBadRequest)
		return
	}

	encryptedPassword, err := EncryptPassword(user.Password)
	if err != nil {
		container.Error(err, http.StatusInternalServerError)
		return
	}
	user.SetPassword("")
	user.SetEncryptedPassword(encryptedPassword)

	if err = container.GetRepository().Create(user); err != nil {
		container.Error(err, http.StatusInternalServerError)
		return
	}
	container.GetResponseWriter().WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(container.GetResponseWriter()).Encode(struct {
		ID   int64
		Link string
	}{
		ID:   user.GetID(),
		Link: fmt.Sprintf("/users/%d", user.GetID()),
	}); err != nil {
		container.Error(err, http.StatusInternalServerError)
	}
}
func (module UserEndpoint) Wrap(handler func(UserEndpointContainer)) func(tiger.Container) {
	return func(c tiger.Container) {
		if container, ok := c.(UserEndpointContainer); !ok {
			c.Error(fmt.Errorf("Container does not implement Userscontainer"), http.StatusInternalServerError)
		} else {
			handler(container)
		}

	}
}
