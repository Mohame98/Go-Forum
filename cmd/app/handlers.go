package main

import (
	"github.com/Mohame98/go-forum/internal/validator"
	"github.com/Mohame98/go-forum/internal/models"
	"net/http"
	"strconv"
	"errors"
	"fmt"
)

// home handles the request for the home page.
// It retrieves the latest threads and renders the home template with the thread data.
func (app *application) home(w http.ResponseWriter, r *http.Request){
	data := app.newTemplateData(r)
	uid := app.currentUserID(r)

	page := 1
    if r.URL.Query().Get("page") != "" {
        var err error
        page, err = strconv.Atoi(r.URL.Query().Get("page"))
        if err != nil || page < 1 { page = 1 }
    }

	perPage := 5
	offset := (page - 1) * perPage
 
	totalThreads, err := app.threads.CountThreads()
	if err != nil { app.serverError(w, r, err); return }
 
	threads, err := app.threads.GetLatestThreads(perPage, offset)
	if err != nil { app.serverError(w, r, err); return }

	if uid != 0 {
        user, err := app.users.GetUserByID(uid)
        if err != nil {app.serverError(w, r, err); return }
		data.User = user
	}
       
	data.Threads = threads
	totalPages := (totalThreads + perPage - 1) / perPage

	var prevLink, nextLink string
	if page > 1 { prevLink = fmt.Sprintf("?page=%d", page-1) }
	if page < totalPages { nextLink = fmt.Sprintf("?page=%d", page+1) }
 
	data.PrevLink = prevLink
	data.NextLink = nextLink
	app.render(w, r, http.StatusOK, "home.tmpl", data)
}

func (app *application) about(w http.ResponseWriter, r *http.Request){
	data := app.newTemplateData(r)
	app.render(w, r, http.StatusOK, "about.tmpl", data)
}

// readThread handles the request to read a specific thread.
// It retrieves messages for the thread and the thread details, then renders the read thread template.
func (app *application) readThread(w http.ResponseWriter, r *http.Request){
	id := app.parseID(w, r)
	data := app.newTemplateData(r)

	threads, err := app.threads.GetThreadByID(id)
	if err != nil { app.serverError(w, r, err); return }

	data.Thread = threads
	app.render(w, r, http.StatusOK, "read-thread.tmpl", data)
}

type threadCreateForm struct {
	Title   string
	validator.Validator
}

type messageCreateForm struct {
	Message   string
	validator.Validator
}

// createThread handles the request to display the create thread form.
// It renders the create thread template without any initial data.
func (app *application) createThread(w http.ResponseWriter, r *http.Request){
	data := app.newTemplateData(r)
	data.Form = threadCreateForm{}
	app.render(w, r, http.StatusOK, "create-thread.tmpl", data)
}

// Inserts a new thread into the database and redirects to the home page.
func (app *application) createThreadPost(w http.ResponseWriter, r *http.Request){
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	err := r.ParseForm()
	if err != nil { app.clientError(w, http.StatusBadRequest); return }

	form := threadCreateForm{ Title:   r.PostForm.Get("title"), }
	form.CheckField(validator.NotBlank(form.Title), "title", "Title cannot be blank")
	form.CheckField(validator.MaxChars(form.Title, 100), "title", "Title cannot be more than 100 characters")
	
	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "create-thread.tmpl", data)
		return
	}
 
	aid := app.currentUserID(r)
	title := r.PostForm.Get("title")

	id, err := app.threads.Insert(title, aid)
	if err != nil { app.serverError(w, r, err); return }

	app.sessionManager.Put(r.Context(), "flash", "Thread created!")
	http.Redirect(w, r, fmt.Sprintf("/read/thread/%v", id), http.StatusSeeOther)
}

// handles the request to display the message posting form for a specific thread.
// It retrieves the thread details and renders the post message template.
func (app *application) newThreadMessage(w http.ResponseWriter, r *http.Request){
	tid := app.parseID(w, r)
	data := app.newTemplateData(r)
	data.Form = messageCreateForm{}
	threads, err := app.threads.GetThreadByID(tid)
	if err != nil { app.serverError(w, r, err); return }

	data.Thread = threads
	app.render(w, r, http.StatusOK, "post-message.tmpl", data)
}

// handles the submission of a new message for a specific thread.
// It inserts the message into the database and redirects to the thread reading page.
func (app *application) newThreadMessagePost(w http.ResponseWriter, r *http.Request){
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	tid := app.parseID(w, r)

	err := r.ParseForm()
	if err != nil { app.clientError(w, http.StatusBadRequest); return }
	
	form := messageCreateForm{ Message: r.PostForm.Get("message"), }
	form.CheckField(validator.NotBlank(form.Message), "message", "Message cannot be blank")
	form.CheckField(validator.MaxChars(form.Message, 1000), "message", "Message cannot be more than 1000 characters")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		
		threads, err := app.threads.GetThreadByID(tid)
		if err != nil { app.serverError(w, r, err); return }

		data.Thread = threads
		app.render(w, r, http.StatusSeeOther, "post-message.tmpl", data)
		return
	}

	aid := app.currentUserID(r)
	message := r.PostForm.Get("message")

	_, err = app.messages.Insert(tid, aid, message)
	if err != nil { app.serverError(w, r, err); return }

	app.sessionManager.Put(r.Context(), "flash", "New Message Posted")
	http.Redirect(w, r, fmt.Sprintf("/read/thread/%v", tid), http.StatusSeeOther)
}

// user handlers register
type userRegisterForm struct {
	User     string
	Email    string
	Password string
	validator.Validator
}

func (app *application) userRegister(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userRegisterForm{}
	app.render(w, r, http.StatusOK, "register.tmpl", data) 
}

func (app *application) userRegisterPost(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	err := r.ParseForm()
	if err != nil { app.clientError(w, http.StatusBadRequest); return }

	form := userRegisterForm{
		User:     r.PostForm.Get("user"),
		Email:    r.PostForm.Get("email"),
		Password: r.PostForm.Get("password"),
	}

	form.CheckField(validator.NotBlank(form.User), "user", "User cannot be blank")
	form.CheckField(validator.NotBlank(form.Email), "email", "Email cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", " Must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "Password cannot be blank")
	form.CheckField(validator.MinChars(form.Password, 8), "password", "Password must be at least 8 characters long")
	form.CheckField(validator.Matches(form.Password, validator.PasswordUpperRX), "password", "Password must contain at least one uppercase letter")
	form.CheckField(validator.Matches(form.Password, validator.PasswordNumberRX), "password", "Password must contain at least one number")

	err = app.users.CheckDuplicateEmail(form.Email)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			form.AddNonFieldError("Email already exists")
			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, r, http.StatusUnprocessableEntity, "register.tmpl", data)
			return
		}
		app.serverError(w, r, err) 
		return
	}

	err = app.users.CheckDuplicateUser(form.User)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateUser) {
			form.AddNonFieldError("User already exists")
			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, r, http.StatusUnprocessableEntity, "register.tmpl", data)
			return
		}
		app.serverError(w, r, err) 
		return
	}

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "register.tmpl", data)
		return
	}
	err = app.users.Insert(form.User, form.Email, form.Password)
	if err != nil { app.serverError(w, r, err); return }

	app.sessionManager.Put(r.Context(), "flash", "Registered successfully, you may log in.")
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

// user handlers login
type userLoginForm struct {
	Email    string
	Password string
	validator.Validator
}

func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userLoginForm{}
	app.render(w, r, http.StatusOK, "login.tmpl", data)
}

func (app *application) userLoginPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil { app.clientError(w, http.StatusBadRequest); return }

	form := userLoginForm{
		Email:    r.PostForm.Get("email"),
		Password: r.PostForm.Get("password"),
	}

	form.CheckField(validator.NotBlank(form.Email), "email", "Email cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "Must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "Password cannot be blank")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "login.tmpl", data)
		return
	}

	id, err := app.users.Authenticate(form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddNonFieldError("Email or password incorrect")
			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, r, http.StatusUnprocessableEntity, "login.tmpl", data)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	err = app.sessionManager.RenewToken(r.Context())
	if err != nil { app.serverError(w, r, err); return }

	app.sessionManager.Put(r.Context(), "authenticatedUserID", id)
	app.sessionManager.Put(r.Context(), "flash", "Logged In")

	path := app.sessionManager.PopString(r.Context(), "redirectPathAfterLogin")
	if path != "" { http.Redirect(w, r, path, http.StatusSeeOther); return }

	http.Redirect(w, r, "/user/account/view", http.StatusSeeOther)
}

// view account
func (app *application) viewAccount(w http.ResponseWriter, r *http.Request){
	data := app.newTemplateData(r)
	uid := app.currentUserID(r)

	user, err := app.users.GetUserByID(uid)
	if err != nil { app.serverError(w, r, err); return }

	data.User = user
    data.CurrentUser = uid 
	app.render(w, r, http.StatusOK, "account.tmpl", data)
}

// user handlers login
type updatePasswordForm struct {
	CurrentPassword 	string
	NewPassword 		string
	NewPasswordConfirm 	string
	validator.Validator
}

// update account
func (app *application) updatePassword(w http.ResponseWriter, r *http.Request){
	uid := app.currentUserID(r)
	data := app.newTemplateData(r)
	data.Form = updatePasswordForm{}
	
	user, err := app.users.GetUserByID(uid)
	if err != nil { app.serverError(w, r, err); return }

	data.User = user
    data.CurrentUser = uid 
	app.render(w, r, http.StatusOK, "update-password.tmpl", data)
}

func (app *application) updatePasswordPost(w http.ResponseWriter, r *http.Request){
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	err := r.ParseForm()
	if err != nil { app.clientError(w, http.StatusBadRequest); return }

	form := updatePasswordForm{
		CurrentPassword: 	r.PostForm.Get("currentPassword"),
		NewPassword: 		r.PostForm.Get("newPassword"),
		NewPasswordConfirm: r.PostForm.Get("newPasswordConfirm"),
	}

	form.CheckField(validator.NotBlank(form.CurrentPassword), "currentPassword", "Current Password cannot be blank")
	form.CheckField(validator.NotBlank(form.NewPassword), "newPassword", "New Password cannot be blank")
	form.CheckField(validator.MinChars(form.NewPassword, 8), "newPassword", "This field must be at least 8 characters long")
	form.CheckField(validator.NotBlank(form.NewPasswordConfirm), "newPasswordConfirm", "New Password Confirm cannot be blank")
	form.CheckField(form.NewPassword == form.NewPasswordConfirm, "newPasswordConfirm", "Passwords do not match")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "update-password.tmpl", data)
		return
	}

	uid := app.currentUserID(r)
	err = app.users.PasswordUpdate(uid, form.CurrentPassword, form.NewPassword)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddFieldError("currentPassword", "Current password is incorrect")
			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, r, http.StatusUnprocessableEntity, "update-password.tmpl", data)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Your password has been updated!")
	http.Redirect(w, r, "/user/account/view", http.StatusSeeOther)
}

func (app *application) userLogoutPost(w http.ResponseWriter, r *http.Request){
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil { app.serverError(w, r, err); return }

	app.sessionManager.Remove(r.Context(), "authenticatedUserID")
	app.sessionManager.Put(r.Context(), "flash", "Logged out")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}