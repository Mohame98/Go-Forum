package main

import (
	"forum/internal/models"
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"time"

	"github.com/justinas/nosurf"
)

// templateData holds data that is passed to templates for rendering.
// It includes information about the current year, threads, messages, and other relevant data.
type templateData struct {
	CurrentYear 	int
	Thread      	*models.Thread   
	Threads     	[]*models.Thread
	Message     	*models.Message    
	Messages    	[]*models.Message
	User        	*models.User
	Users       	[]*models.User
	Form        	any
	Flash       	string
	IsAuthenticated bool  
	CurrentUser 	int
	CSRFToken   	string
	PrevLink 		string
    NextLink 		string	
}

// humanDate formats a given time.Time into a readable string.
// The output format is "02 Jan 2006 at 15:04".
func humanDate(t time.Time) string {
	return t.Format("02 Jan 2006 at 15:04")
}

// functions is a map of template functions available for use in templates.
// It includes functions for formatting dates and other utilities.
var functions = template.FuncMap{
	"humanDate": humanDate,
}

// newTemplateData initializes and returns a templateData struct populated
// with the current year. It is used for preparing template data before rendering.
func (app *application) newTemplateData(r *http.Request) templateData {
	return templateData{
		CurrentYear: 		time.Now().Year(),
		Flash: 				app.sessionManager.PopString(r.Context(), "flash"),	
		IsAuthenticated: 	app.isAuthenticated(r),	
		CSRFToken: 			nosurf.Token(r),
	}
}

// newTemplateCache creates a map of template names to parsed templates.
// It loads templates from the specified directory and parses them into the cache.
// Returns the cache and any error 
func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}
	pages, err := filepath.Glob("./ui/html/pages/*.tmpl")
	if err != nil { return nil, err }

	for _, page := range pages {
		name := filepath.Base(page)
		ts, err := template.New(name).Funcs(functions).ParseFiles("./ui/html/base.tmpl")
		if err != nil { return nil, err }
	
		ts, err = ts.ParseGlob("./ui/html/partials/*.tmpl")
		if err != nil { return nil, err }
	
		ts, err = ts.ParseFiles(page)
		if err != nil { return nil, err }
	
		cache[name] = ts
	}
	return cache, nil
}

// render executes a template and writes it to the HTTP response. If the template
// is not found, it responds with Internal Server Error.
func (app *application) render(
	w http.ResponseWriter, 
	r *http.Request, 
	status int, 
	page string,
	data templateData,
) {
	ts, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		app.serverError(w, r, err)
		return
	}

	buf := new(bytes.Buffer)
	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil { app.serverError(w, r, err)
		return
	}

	w.WriteHeader(status)
	buf.WriteTo(w)
}