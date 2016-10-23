package smartsnippets_test

//import (
//	"testing"

//	app "github.com/Mparaiso/code-snippet-manager-go"
//	"github.com/Mparaiso/expect-go"
//	"github.com/Mparaiso/simple-validation-go"
//	"google.golang.org/appengine/aetest"
//)

//func TestSnippetValidator(t *testing.T) {
//	context, done, err := aetest.NewContext()
//	expect.Expect(t, err, nil)
//	defer done()
//	categoryRepository := app.NewCategoryRepository(context)
//	category := &app.Category{Title: "PHP", Description: "PHP language"}
//	err = categoryRepository.Create(category)
//	phpCategory := &app.Category{}
//	err = categoryRepository.FindByID(category.ID, phpCategory)
//	expect.Expect(t, err, nil)
//	expect.Expect(t, phpCategory.Title, "PHP")
//	expect.Expect(t, phpCategory.ID > 0, true)
//	snippet := &app.Snippet{Title: "Hello PHP", Content: `<?php echo "Hello World!";`}
//	snippetValidator := app.NewSnippetValidator(&app.DefaultExistingEntityValidatorProvider{categoryRepository})
//	err = snippetValidator.Validate(snippet)
//	expect.Expect(t, err != nil, true)
//	_, ok := err.(*validation.ConcreteError).GetErrors()["CategoryID"]
//	expect.Expect(t, ok, true)
//}
//func TestCategoryValidator(t *testing.T) {
//	context, done, err := aetest.NewContext()
//	expect.Expect(t, err, nil)
//	defer done()
//	t.Run("Valid Category", func(t *testing.T) {
//		categoryRepository := app.NewCategoryRepository(context)
//		category := &app.Category{Title: "PHP", Description: "PHP language"}
//		categoryValidator := &app.CategoryValidator{&app.DefaultUniqueEntityValidatorProvider{categoryRepository}}
//		err = categoryValidator.Validate(category)
//		expect.Expect(t, err, nil)
//		err = categoryRepository.Create(category)
//		expect.Expect(t, err, nil)
//		phpCategory := &app.Category{}
//		err = categoryRepository.FindByID(category.ID, phpCategory)
//		expect.Expect(t, err, nil)

//	})
//	t.Run("Invalid duplicate Category", func(t *testing.T) {
//		categoryRepository := app.NewCategoryRepository(context)
//		category := &app.Category{Title: "PHP", Description: "PHP language"}
//		categoryValidator := &app.CategoryValidator{&app.DefaultUniqueEntityValidatorProvider{categoryRepository}}
//		err = categoryValidator.Validate(category)
//		expect.Expect(t, err != nil, true)
//		t.Log(err)
//	})

//}
