package main

import (
	"SnippetAppBook/internal/models"
	"SnippetAppBook/internal/validator"
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
)

type userLoginForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

type userSignupForm struct {
	Name                string `form:"name"`
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

type SnippetCreateForm struct {
	Title               string `form:"title"`
	Content             string `form:"content"`
	Expires             int    `form:"expires"`
	validator.Validator `form:"-"`
}

func (app *Application) home(w http.ResponseWriter, r *http.Request) {

	snippets, err := app.snippets.Latest()
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r)
	data.Snippets = snippets
	app.render(w, http.StatusOK, "home.tmpl", data)

}

func (app *Application) view(w http.ResponseWriter, r *http.Request) {

	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil || id < 0 {
		app.notFound(w)
		return
	}
	//w.Write([]byte(app.snippets.Get(1))

	snippet, err := app.snippets.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}
	//

	data := app.newTemplateData(r)
	data.Snippet = snippet

	app.render(w, http.StatusOK, "view.tmpl", data)

}

func (app *Application) test(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("test"))

}

func (app *Application) snippetCreate(w http.ResponseWriter, r *http.Request) {

	data := app.newTemplateData(r)
	data.Form = SnippetCreateForm{
		Expires: 365,
	}
	app.render(w, http.StatusOK, "create.tmpl", data)

}

func (app *Application) snippetCreatePost(w http.ResponseWriter, r *http.Request) {

	var form SnippetCreateForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Title), "title", "this cannot be blank")
	form.CheckField(validator.MaxChars(form.Title, 100), "title", "this cannot be more than 100 characters long")
	form.CheckField(validator.NotBlank(form.Content), "content", "this cannot be blank")
	form.CheckField(validator.PermittedVal(form.Expires, 1, 7, 365), "expires", "must be 1 or 7 or 365")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "create.tmpl", data)
		return
	}

	id, err := app.snippets.Insert(form.Title, form.Content, form.Expires)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Snippet created")

	http.Redirect(w, r, fmt.Sprintf("/view/%d", id), http.StatusSeeOther)

}

func (app *Application) signUp(w http.ResponseWriter, r *http.Request) {

	data := app.newTemplateData(r)
	data.Form = userSignupForm{}
	//w.Write([]byte(data.))

	app.render(w, http.StatusOK, "signup.tmpl", data)

}

func (app *Application) signUpPost(w http.ResponseWriter, r *http.Request) {

	var form userSignupForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Name), "name", "cannot be blanc")
	form.CheckField(validator.NotBlank(form.Email), "email", "cannot be blanc")
	form.CheckField(validator.NotBlank(form.Password), "password", "cannot be blanc")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "must be correct email")
	form.CheckField(validator.MaxChars(form.Password, 8), "password", "must be at least 8 long")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "signup.tmpl", data)
		return
	}

	err = app.users.Insert(form.Name, form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			form.AddFieldError("email", "Email address is already in use")
			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "signup.tmpl", data)
		} else {
			app.serverError(w, err)
		}
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Your signup was successful. Please log in.")

	http.Redirect(w, r, "/login", http.StatusSeeOther)

}

func (app *Application) login(w http.ResponseWriter, r *http.Request) {

	data := app.newTemplateData(r)
	data.Form = userLoginForm{}
	app.render(w, http.StatusOK, "login.tmpl", data)

}

func (app *Application) loginPost(w http.ResponseWriter, r *http.Request) {

	var form userLoginForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Email), "email", "cannot be blanc")
	form.CheckField(validator.NotBlank(form.Password), "password", "cannot be blanc")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "must be correct email")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "login.tmpl", data)
		return
	}
	id, err := app.users.Authenticate(form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddNonFieldError("invalid password or email")
			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "login.tmpl", data)
			return
		} else {
			app.serverError(w, err)
		}
		return
	}
	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.sessionManager.Put(r.Context(), "authUserId", id)
	http.Redirect(w, r, "/create", http.StatusSeeOther)
}

func (app *Application) logout(w http.ResponseWriter, r *http.Request) {

	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.sessionManager.Remove(r.Context(), "authUserId")
	app.sessionManager.Put(r.Context(), "flash", "you been loged out")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}
