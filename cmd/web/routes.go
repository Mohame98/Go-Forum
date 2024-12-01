package main

import (
	"github.com/justinas/alice"	
	"net/http"
)

func (app *application) routes() http.Handler{
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./ui/static"))
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

	dynamic := alice.New(app.sessionManager.LoadAndSave, noSurf) // No auth middleware
	protected := dynamic.Append(app.requireAuthentication) // Auth middleware

	mux.Handle("GET /{$}", dynamic.ThenFunc(app.home))
	mux.Handle("GET /about", dynamic.ThenFunc(app.about))
	mux.Handle("GET /read/thread/{id}", protected.ThenFunc(app.readThread))

	// threads/thread messages
	mux.Handle("GET /create/thread", protected.ThenFunc(app.createThread))
	mux.Handle("POST /create/thread", protected.ThenFunc(app.createThreadPost))
	
	mux.Handle("GET /thread/newmessage/{id}", protected.ThenFunc(app.newThreadMessage))
	mux.Handle("POST /thread/newmessage/{id}", protected.ThenFunc(app.newThreadMessagePost))

	// user
	mux.Handle("GET /user/register", dynamic.ThenFunc(app.userRegister))
	mux.Handle("POST /user/register", dynamic.ThenFunc(app.userRegisterPost))

	mux.Handle("GET /user/login", dynamic.ThenFunc(app.userLogin))
	mux.Handle("POST /user/login", dynamic.ThenFunc(app.userLoginPost))

	mux.Handle("GET /user/account/view", protected.ThenFunc(app.viewAccount))

	mux.Handle("GET /account/password/update", protected.ThenFunc(app.updatePassword))
	mux.Handle("POST /account/password/update", protected.ThenFunc(app.updatePasswordPost))

	mux.Handle("POST /user/logout", protected.ThenFunc(app.userLogoutPost))

	standard := alice.New(app.recoverPanic, app.logRequest, commonHeaders)
	// applied middleware
	return standard.Then(mux)
}
