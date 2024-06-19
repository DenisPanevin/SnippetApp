package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/go-playground/form/v4"
	"github.com/justinas/nosurf"
	"net/http"
	"runtime/debug"
	"time"
)

func (app *Application) serverError(w http.ResponseWriter, err error) {

	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.errLog.Output(2, trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *Application) clientError(w http.ResponseWriter, statusInt int) {
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

}

func (app *Application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)

}

func (app *Application) render(w http.ResponseWriter, status int, page string, data *templateData) {

	ts, ok := app.templateCashe[page]

	if !ok {

		err := fmt.Errorf("the template %s does not exist", page)
		app.serverError(w, err)

		return
	}

	buf := new(bytes.Buffer)
	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(w, err)
		return
	}
	w.WriteHeader(status)

	buf.WriteTo(w)
}

func (app *Application) newTemplateData(r *http.Request) *templateData {
	return &templateData{
		Year:      time.Now().Year(),
		Flash:     app.sessionManager.PopString(r.Context(), "flash"),
		IsAuth:    app.IsAuth(r),
		CSRFToken: nosurf.Token(r),
	}
}

func (app *Application) decodePostForm(r *http.Request, dst any) error {
	err := r.ParseForm()
	if err != nil {
		return err
	}
	err = app.formDecoder.Decode(dst, r.PostForm)
	if err != nil {
		var invalidDecode *form.InvalidDecoderError
		if errors.As(err, &invalidDecode) {
			panic(err)
		}
		return err
	}
	return nil
}

func (app *Application) IsAuth(r *http.Request) bool {
	//return app.sessionManager.Exists(r.Context(), "authUserId")
	isAuth, ok := r.Context().Value(IsAuthContKey).(bool)
	if !ok {
		return false
	}
	return isAuth
}
