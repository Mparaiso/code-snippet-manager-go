package smartsnippets

import (
	"fmt"

	"github.com/Mparaiso/simple-validation-go"
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
func (validator *SnippetValidator) Validate(snippet *Snippet) error {
	errors := validation.NewConcreteError()
	validation.StringNotEmptyValidator("Title", snippet.Title, errors)
	validation.StringLengthValidator("Title", snippet.Title, 8, 127, errors)
	validation.StringLengthValidator("Content", snippet.Content, 8, 2048, errors)
	if !validation.StringEmpty(snippet.Description) {
		validation.StringLengthValidator("Description", snippet.Description, 5, 256, errors)
	}
	validator.ExistingEntityValidator("CategoryID", "Category", map[string]interface{}{"ID": snippet.CategoryID}, errors)
	if errors.HasErrors() {
		return errors
	}
	return nil
}

type CategoryValidator struct {
	UniqueEntityValidatorProvider
}

func (validator CategoryValidator) Validate(category *Category) error {
	errors := validation.NewConcreteError()
	validation.StringNotEmptyValidator("Title", category.Title, errors)
	validation.StringLengthValidator("Title", category.Title, 1, 64, errors)
	validator.UniqueEntityValidatorProvider.UniqueEntityValidator("Title", map[string]interface{}{"Title": category.Title}, errors)
	validation.StringNotEmptyValidator("Description", category.Description, errors)
	validation.StringLengthValidator("Description", category.Description, 5, 127, errors)
	if errors.HasErrors() {
		return errors
	}
	return nil
}

type UserValidator struct {
	UniqueEntityValidatorProvider
}

func (validator UserValidator) Validate(user *User) error {
	errors := validation.NewConcreteError()
	validation.StringNotEmptyValidator("Nickname", user.Nickname, errors)
	validator.UniqueEntityValidator("Nickname", map[string]interface{}{"Nickname": user.Nickname}, errors)
	validation.EmailValidator("Email", user.Email, errors)
	validation.StringNotEmptyValidator("Password", user.Password, errors)
	validation.StringLengthValidator("Password", user.Password, 7, 126, errors)
	if errors.HasErrors() {
		return errors
	}
	return nil
}

type DefaultUniqueEntityValidatorProvider struct {
	Repository
}

func (provider DefaultUniqueEntityValidatorProvider) UniqueEntityValidator(field string, values map[string]interface{}, errors validation.Error) {
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

func (provider DefaultExistingEntityValidatorProvider) ExistingEntityValidator(field string, kind string, values map[string]interface{}, errors validation.Error) {
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
