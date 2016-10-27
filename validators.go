package smartsnippets

import (
	"fmt"

	validator "github.com/Mparaiso/tiger-go-framework/validator"
)

type BasicEntity struct {
	ID int64
}

func (entity BasicEntity) GetID() int64    { return entity.ID }
func (entity *BasicEntity) SetID(id int64) { entity.ID = id }

type SnippetValidator struct {
	ExistingEntityValidatorProvider
}

func NewSnippetValidator(ExistingCategoryValidatorProvider ExistingEntityValidatorProvider) *SnippetValidator {
	return &SnippetValidator{ExistingCategoryValidatorProvider}
}
func (v *SnippetValidator) Validate(snippet *Snippet) error {
	errors := validator.NewConcreteError()
	validator.StringNotEmptyValidator("Title", snippet.Title, errors)
	validator.StringLengthValidator("Title", snippet.Title, 8, 127, errors)
	validator.StringLengthValidator("Content", snippet.Content, 8, 2048, errors)
	if !validator.StringEmpty(snippet.Description) {
		validator.StringLengthValidator("Description", snippet.Description, 5, 256, errors)
	}
	v.ExistingEntityValidator("CategoryID", "Category", map[string]interface{}{"ID": snippet.CategoryID}, errors)
	if errors.HasErrors() {
		return errors
	}
	return nil
}

type CategoryValidator struct {
	UniqueEntityValidatorProvider
}

func (v CategoryValidator) Validate(category *Category) error {
	errors := validator.NewConcreteError()
	validator.StringNotEmptyValidator("Title", category.Title, errors)
	validator.StringLengthValidator("Title", category.Title, 1, 64, errors)
	v.UniqueEntityValidatorProvider.UniqueEntityValidator("Title", map[string]interface{}{"Title": category.Title}, errors)
	validator.StringNotEmptyValidator("Description", category.Description, errors)
	validator.StringLengthValidator("Description", category.Description, 5, 127, errors)
	if errors.HasErrors() {
		return errors
	}
	return nil
}

// UserValidator validates a *User model
type UserValidator struct {
	UniqueEntityValidatorProvider
}

func (v UserValidator) Validate(user *User) error {
	errors := validator.NewConcreteError()
	validator.StringNotEmptyValidator("Nickname", user.Nickname, errors)
	v.UniqueEntityValidator("Nickname", map[string]interface{}{"Nickname": user.Nickname}, errors)
	validator.EmailValidator("Email", user.Email, errors)
	validator.StringNotEmptyValidator("Password", user.Password, errors)
	validator.StringLengthValidator("Password", user.Password, 7, 126, errors)
	if errors.HasErrors() {
		return errors
	}
	return nil
}

type DefaultUniqueEntityValidatorProvider struct {
	Repository
}

func (provider DefaultUniqueEntityValidatorProvider) UniqueEntityValidator(field string, values map[string]interface{}, errors validator.Error) {
	query := map[string]interface{}{}
	for key, value := range values {
		query[key+"="] = value
	}
	count, err := provider.Repository.Count(Query{Query: query, Limit: 1})
	if err != nil {
		errors.Append(field, err.Error())
		return

	}
	if count != 0 {
		errors.Append(field, "Should be unique")
	}

}

type DefaultExistingEntityValidatorProvider struct {
	Repository
}

func NewDefaultExistingEntityValidatorProvider(repository Repository) *DefaultExistingEntityValidatorProvider {
	return &DefaultExistingEntityValidatorProvider{repository}
}

func (provider DefaultExistingEntityValidatorProvider) ExistingEntityValidator(field string, kind string, values map[string]interface{}, errors validator.Error) {
	query := map[string]interface{}{}
	for key, value := range values {
		query[key+"="] = value
	}
	count, err := provider.Repository.Count(Query{Query: query, Limit: 1})
	if err != nil {
		errors.Append(field, err.Error())
		return
	}
	if count == 0 {
		errors.Append(field, fmt.Sprintf("%s with fields matching %v does not exist", kind, values))
	}
}
