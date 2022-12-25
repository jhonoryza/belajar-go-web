package main

import (
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"snippetbox.labkita.my.id/internal/models"
	"snippetbox.labkita.my.id/internal/validator"
	"strconv"
)

func (app *application) home(resp http.ResponseWriter, req *http.Request) {
	snippets, err := app.snippets.Latest()
	if err != nil {
		app.serverError(resp, err)
		return
	}
	data := app.newTemplateData(req)
	data.Snippets = snippets
	app.render(resp, http.StatusOK, "home.tmpl", data)
}

func (app *application) snippetList(resp http.ResponseWriter, req *http.Request) {
	snippets, err := app.snippets.Latest()
	if err != nil {
		app.serverError(resp, err)
		return
	}
	for _, snippet := range snippets {
		fmt.Fprintf(resp, "%+v\n", snippet)
	}
}

func (app *application) snippetView(resp http.ResponseWriter, req *http.Request) {
	//validation id
	params := httprouter.ParamsFromContext(req.Context())
	//id, err := strconv.Atoi(req.URL.Query().Get("id"))
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil || id < 1 {
		app.notFound(resp)
		return
	}

	//query by id
	snippet, err := app.snippets.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(resp)
		} else {
			app.serverError(resp, err)
		}
		return
	}

	data := app.newTemplateData(req)
	data.Snippet = snippet

	app.render(resp, http.StatusOK, "view.tmpl", data)
}

func (app *application) snippetCreateForm(resp http.ResponseWriter, req *http.Request) {
	data := app.newTemplateData(req)

	data.Form = snippetCreateForm{
		Expires: 365,
	}

	app.render(resp, http.StatusOK, "create.tmpl", data)
}

type snippetCreateForm struct {
	Title               string `form:"title"`
	Content             string `form:"content"`
	Expires             int    `form:"expires"`
	validator.Validator `form:"-"`
}

func (app *application) snippetCreate(resp http.ResponseWriter, req *http.Request) {
	// cek bad request
	var form snippetCreateForm
	err := app.decodePostForm(req, &form)
	if err != nil {
		app.clientError(resp, http.StatusBadRequest)
		return
	}

	/** validation snippets: https://www.alexedwards.net/blog/validation-snippets-for-go */
	form.CheckField(validator.IsNotBlank(form.Title), "title", "this field cannot be blank")
	form.CheckField(validator.MaxChars(form.Title, 100), "title", "This field cannot be more than 100 characters long")
	form.CheckField(validator.IsNotBlank(form.Content), "content", "this field cannot be blank")
	form.CheckField(validator.PermittedInt(form.Expires, 1, 7, 365), "expires", "This field must equal 1, 7 or 365")

	if !form.IsValid() {
		data := app.newTemplateData(req)
		data.Form = form
		app.render(resp, http.StatusUnprocessableEntity, "create.tmpl", data)
		return
	}

	// create data
	id, err := app.snippets.Insert(form.Title, form.Content, form.Expires)
	if err != nil {
		app.serverError(resp, err)
		return
	}

	// flash message
	app.sessionManager.Put(req.Context(), "flash", "Snippet successfully created!")

	// redirection
	http.Redirect(resp, req, fmt.Sprintf("/snippets/view/%d", id), http.StatusSeeOther)
}

type userSignupForm struct {
	Name                string `form:"name"`
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

func (app *application) userSignupForm(resp http.ResponseWriter, req *http.Request) {
	data := app.newTemplateData(req)
	data.Form = userSignupForm{}
	app.render(resp, http.StatusOK, "signup.tmpl", data)
}
func (app *application) userSignup(resp http.ResponseWriter, req *http.Request) {
	var form userSignupForm
	err := app.decodePostForm(req, &form)
	if err != nil {
		app.clientError(resp, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.IsNotBlank(form.Name), "name", "this field cannot be blank")
	form.CheckField(validator.IsNotBlank(form.Email), "email", "this field cannot be blank")
	form.CheckField(validator.IsNotBlank(form.Password), "password", "this field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "this field must be a valid email address")
	form.CheckField(validator.MinChars(form.Password, 8), "password", "this field must be at least 8 char")

	if !form.IsValid() {
		data := app.newTemplateData(req)
		data.Form = form
		app.render(resp, http.StatusUnprocessableEntity, "signup.tmpl", data)
		return
	}

	err = app.users.Insert(form.Name, form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			form.AddFieldError("email", "Email address is already in use")
			data := app.newTemplateData(req)
			data.Form = form
			app.render(resp, http.StatusUnprocessableEntity, "signup.tmpl", data)
		} else {
			app.serverError(resp, err)
		}
		return
	}

	app.sessionManager.Put(req.Context(), "flash", "Your signup was successful. Please log in.")

	http.Redirect(resp, req, "/user/login", http.StatusSeeOther)
}

type userLoginForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

func (app *application) userLoginForm(resp http.ResponseWriter, req *http.Request) {
	data := app.newTemplateData(req)
	data.Form = userLoginForm{}
	app.render(resp, http.StatusOK, "login.tmpl", data)
}

func (app *application) userLogin(resp http.ResponseWriter, req *http.Request) {
	var form userLoginForm
	err := app.decodePostForm(req, &form)
	if err != nil {
		app.clientError(resp, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.IsNotBlank(form.Email), "email", "this field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "this field must be valid email address")
	form.CheckField(validator.IsNotBlank(form.Password), "password", "this field cannot be blank")

	if !form.IsValid() {
		data := app.newTemplateData(req)
		data.Form = form
		app.render(resp, http.StatusUnprocessableEntity, "login.tmpl", data)
		return
	}

	id, err := app.users.Authenticate(form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddNonFieldError("Email or password is incorrect")
			data := app.newTemplateData(req)
			data.Form = form
			app.render(resp, http.StatusUnprocessableEntity, "login.tmpl", data)
		} else {
			app.serverError(resp, err)
		}
		return
	}

	err = app.sessionManager.RenewToken(req.Context())
	if err != nil {
		app.serverError(resp, err)
		return
	}

	app.sessionManager.Put(req.Context(), "authenticatedUserID", id)

	http.Redirect(resp, req, "/snippets/create", http.StatusSeeOther)
}

func (app *application) userLogout(resp http.ResponseWriter, req *http.Request) {
	err := app.sessionManager.RenewToken(req.Context())
	if err != nil {
		app.serverError(resp, err)
		return
	}
	app.sessionManager.Remove(req.Context(), "authenticatedUserID")
	app.sessionManager.Put(req.Context(), "flash", "You've been logged out successfully!")
	http.Redirect(resp, req, "/", http.StatusSeeOther)
}
