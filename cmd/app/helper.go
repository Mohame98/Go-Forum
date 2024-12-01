package main

import (
	"net/http"
	"strconv"
)

// serverError logs the error and sends Internal Server Error response
// with a generic message to the client.
func (app *application) serverError(w http.ResponseWriter, r *http.Request, err error){
	var ( 
        method = r.Method
		uri = r.URL.RequestURI()
	)
	app.logger.Error(err.Error(), "method", method, "uri", uri)
	http.Error(
		w,
		http.StatusText(http.StatusInternalServerError),
		http.StatusInternalServerError,
	)
}

// clientError sends an HTTP response with the specified status code and
// corresponding status text message.
func (app *application) clientError(w http.ResponseWriter, status int){
	http.Error(w, http.StatusText(status), status)
}

// parseID extracts and converts the 'id' from the request path to an integer.
// If invalid, it responds with a client error if not returns the id
func (app *application) parseID(w http.ResponseWriter, r *http.Request) int {
	id, err := strconv.Atoi(r.PathValue("id"))
    if err != nil || id < 1 { app.clientError(w, http.StatusBadRequest); return 0 }
    return id
}

func (app *application) isAuthenticated(r *http.Request) bool {
	return app.sessionManager.Exists(r.Context(), "authenticatedUserID")
}

func (app *application) currentUserID(r *http.Request) int {
    userID := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")
    return userID
}