package webgo

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
)

// ErrorData used to render the error page
type ErrorData struct {
	ErrCode        int
	ErrDescription string
}

// dOutput is the standard/valid output wrapped in `{data: <payload>, status: <http response status>}`
type dOutput struct {
	Data   interface{} `json:"data"`
	Status int         `json:"status"`
}

// errOutput is the error output wrapped in `{errors:<errors>, status: <http response status>}`
type errOutput struct {
	Errors interface{} `json:"errors"`
	Status int         `json:"status"`
}

const (
	// HeaderContentType is the key for mentioning the response header content type
	HeaderContentType = "Content-Type"
	// JSONContentType is the MIME type when the response is JSON
	JSONContentType = "application/json"
	// HTMLContentType is the MIME type when the response is HTML
	HTMLContentType = "text/html; charset=UTF-8"

	// ErrInternalServer to send when there's an internal server error
	ErrInternalServer = "Internal server error"
)

// SendHeader is used to send only a response header, i.e no response body
func SendHeader(w http.ResponseWriter, rCode int) {
	w.WriteHeader(rCode)
}

func crwAsserter(w http.ResponseWriter, rCode int) http.ResponseWriter {
	if crw, ok := w.(*customResponseWriter); ok {
		crw.statusCode = rCode
		return crw
	}

	return newCRW(w, rCode)
}

// Send sends a completely custom response without wrapping in the
// `{data: <data>, status: <int>` struct
func Send(w http.ResponseWriter, contentType string, data interface{}, rCode int) {
	w = crwAsserter(w, rCode)

	w.Header().Set(HeaderContentType, contentType)
	_, err := fmt.Fprint(w, data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(ErrInternalServer))
		LOGHANDLER.Error(err)
	}
}

// SendResponse is used to respond to any request (JSON response) based on the code, data etc.
func SendResponse(w http.ResponseWriter, data interface{}, rCode int) {
	w = crwAsserter(w, rCode)
	w.Header().Add(HeaderContentType, JSONContentType)
	err := json.NewEncoder(w).Encode(dOutput{Data: data, Status: rCode})
	if err != nil {
		/*
			In case of encoding error, send "internal server error" and
			log the actual error.
		*/
		R500(w, ErrInternalServer)
		LOGHANDLER.Error(err)
	}
}

// SendError is used to respond to any request with an error
func SendError(w http.ResponseWriter, data interface{}, rCode int) {
	w = crwAsserter(w, rCode)
	w.Header().Add(HeaderContentType, JSONContentType)
	err := json.NewEncoder(w).Encode(errOutput{data, rCode})
	if err != nil {
		/*
			In case of encoding error, send "internal server error" and
			log the actual error.
		*/
		R500(w, ErrInternalServer)
		LOGHANDLER.Error(err)
	}
}

// Render is used for rendering templates (HTML)
func Render(w http.ResponseWriter, data interface{}, rCode int, tpl *template.Template) {
	w = crwAsserter(w, rCode)

	// In case of HTML response, setting appropriate header type for text/HTML response
	w.Header().Set(HeaderContentType, HTMLContentType)

	// Rendering an HTML template with appropriate data
	err := tpl.Execute(w, data)
	if err != nil {
		Send(w, "text/plain", ErrInternalServer, http.StatusInternalServerError)
		LOGHANDLER.Error(err.Error())
	}
}

// Render404 - used to render a 404 page
func Render404(w http.ResponseWriter, tpl *template.Template) {
	Render(w, ErrorData{
		http.StatusNotFound,
		"Sorry, the URL you requested was not found on this server... Or you're lost :-/",
	},
		http.StatusNotFound,
		tpl,
	)
}

// R200 - Successful/OK response
func R200(w http.ResponseWriter, data interface{}) {
	SendResponse(w, data, http.StatusOK)
}

// R201 - New item created
func R201(w http.ResponseWriter, data interface{}) {
	SendResponse(w, data, http.StatusCreated)
}

// R204 - empty, no content
func R204(w http.ResponseWriter) {
	SendHeader(w, http.StatusNoContent)
}

// R302 - Temporary redirect
func R302(w http.ResponseWriter, data interface{}) {
	SendResponse(w, data, http.StatusFound)
}

// R400 - Invalid request, any incorrect/erraneous value in the request body
func R400(w http.ResponseWriter, data interface{}) {
	SendError(w, data, http.StatusBadRequest)
}

// R403 - Unauthorized access
func R403(w http.ResponseWriter, data interface{}) {
	SendError(w, data, http.StatusForbidden)
}

// R404 - Resource not found
func R404(w http.ResponseWriter, data interface{}) {
	SendError(w, data, http.StatusNotFound)
}

// R406 - Unacceptable header. For any error related to values set in header
func R406(w http.ResponseWriter, data interface{}) {
	SendError(w, data, http.StatusNotAcceptable)
}

// R451 - Resource taken down because of a legal request
func R451(w http.ResponseWriter, data interface{}) {
	SendError(w, data, http.StatusUnavailableForLegalReasons)
}

// R500 - Internal server error
func R500(w http.ResponseWriter, data interface{}) {
	SendError(w, data, http.StatusInternalServerError)
}
